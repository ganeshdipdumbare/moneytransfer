package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"moneytransfer/internal/service"
	"moneytransfer/internal/transfer"

	"github.com/gin-gonic/gin"
)

// BulkTransferResponse represents the structure of a successful bulk transfer response
type BulkTransferResponse struct {
	Message string `json:"message"`
}

// BulkTransferFileContent represents the structure of the JSON file content
type BulkTransferFileContent struct {
	OrganizationName string           `json:"organization_name" validate:"required"`
	OrganizationBIC  string           `json:"organization_bic" validate:"required"`
	OrganizationIBAN string           `json:"organization_iban" validate:"required"`
	CreditTransfers  []CreditTransfer `json:"credit_transfers" validate:"required"`
}

// CreditTransfer represents the structure of a credit transfer
type CreditTransfer struct {
	Amount           string `json:"amount" validate:"required"`
	CounterpartyName string `json:"counterparty_name" validate:"required"`
	CounterpartyBIC  string `json:"counterparty_bic" validate:"required"`
	CounterpartyIBAN string `json:"counterparty_iban" validate:"required"`
	Description      string `json:"description" validate:"required"`
}

// BulkTransfer godoc
// @Summary Perform a bulk transfer
// @Description Transfer money from one account to multiple accounts using a file upload
// @Tags transfers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "JSON file containing bulk transfer details"
// @Success 201 {object} BulkTransferResponse
// @Failure 400 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transfers [post]
func (api *apiDetails) BulkTransfer(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		createErrorResponse(c, http.StatusBadRequest, "Error retrieving the file")
		return
	}
	defer file.Close()

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		createErrorResponse(c, http.StatusBadRequest, "Error reading file content")
		return
	}

	var bulkTransferContent BulkTransferFileContent
	err = json.Unmarshal(fileContent, &bulkTransferContent)
	if err != nil {
		createErrorResponse(c, http.StatusBadRequest, "Error parsing JSON content")
		return
	}

	// Validate the bulk transfer content using the validator
	if err := validate.Struct(bulkTransferContent); err != nil {
		createErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	// Convert BulkTransferFileContent to service.BulkTransferRequest
	request := service.BulkTransferRequest{
		OrganizationName: bulkTransferContent.OrganizationName,
		OrganizationBIC:  bulkTransferContent.OrganizationBIC,
		OrganizationIBAN: bulkTransferContent.OrganizationIBAN,
		Transfers:        make([]transfer.Transfer, len(bulkTransferContent.CreditTransfers)),
	}

	for i, ct := range bulkTransferContent.CreditTransfers {
		amount, err := parseAmount(ct.Amount)
		if err != nil {
			createErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid amount for transfer to %s: %v", ct.CounterpartyName, err))
			return
		}
		request.Transfers[i] = transfer.Transfer{
			AmountCents:      amount,
			CounterpartyName: ct.CounterpartyName,
			CounterpartyBIC:  ct.CounterpartyBIC,
			CounterpartyIBAN: ct.CounterpartyIBAN,
			Description:      ct.Description,
		}
	}

	err = api.service.BulkTransfer(c.Request.Context(), request)
	if err != nil {
		if err.Error() == "insufficient funds" {
			createErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		} else {
			createErrorResponse(c, http.StatusInternalServerError, "Error processing bulk transfer")
		}
		return
	}

	c.JSON(http.StatusCreated, BulkTransferResponse{
		Message: "Bulk transfer processed successfully",
	})
}

// parseAmount converts a string amount to int64 cents
func parseAmount(amount string) (int64, error) {
	if amount == "" {
		return 0, fmt.Errorf("amount cannot be empty")
	}

	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount format")
	}

	var cents int64
	var err error

	switch len(parts) {
	case 1:
		// No decimal point
		cents, err = strconv.ParseInt(parts[0]+"00", 10, 64)
	case 2:
		intPart := parts[0]
		decPart := parts[1]

		switch len(decPart) {
		case 1:
			cents, err = strconv.ParseInt(intPart+decPart+"0", 10, 64)
		case 2:
			cents, err = strconv.ParseInt(intPart+decPart, 10, 64)
		default:
			return 0, fmt.Errorf("invalid decimal places")
		}
	}

	if err != nil {
		return 0, fmt.Errorf("error parsing amount: %v", err)
	}

	if cents < 0 {
		return 0, fmt.Errorf("amount cannot be negative")
	}

	return cents, nil
}
