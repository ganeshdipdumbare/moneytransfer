package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"moneytransfer/mock"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestBulkTransfer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		fileContent        string
		setupMock          func(*mock.TransferServiceMock)
		expectedStatusCode int
		expectedResponse   interface{}
	}{
		{
			name: "Successful bulk transfer",
			fileContent: `{
				"organization_name": "Test Org",
				"organization_bic": "TESTBIC1",
				"organization_iban": "TEST123456789",
				"credit_transfers": [
					{
						"amount": "100.50",
						"counterparty_name": "John Doe",
						"counterparty_bic": "JOHNDOEBIC",
						"counterparty_iban": "JOHNDOE987654321",
						"description": "Test transfer"
					}
				]
			}`,
			setupMock: func(mockService *mock.TransferServiceMock) {
				mockService.EXPECT().
					BulkTransfer(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: BulkTransferResponse{
				Message: "Bulk transfer processed successfully",
			},
		},
		{
			name: "Invalid JSON content",
			fileContent: `{
				"invalid": "json"
			}`,
			setupMock:          func(mockService *mock.TransferServiceMock) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: ErrorResponse{
				Message: "invalid request",
			},
		},
		{
			name: "Invalid amount",
			fileContent: `{
				"organization_name": "Test Org",
				"organization_bic": "TESTBIC1",
				"organization_iban": "TEST123456789",
				"credit_transfers": [
					{
						"amount": "invalid",
						"counterparty_name": "John Doe",
						"counterparty_bic": "JOHNDOEBIC",
						"counterparty_iban": "JOHNDOE987654321",
						"description": "Test transfer"
					}
				]
			}`,
			setupMock:          func(mockService *mock.TransferServiceMock) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: ErrorResponse{
				Message: "Invalid amount for transfer to John Doe: error parsing amount: strconv.ParseInt: parsing \"invalid00\": invalid syntax",
			},
		},
		{
			name: "Insufficient funds",
			fileContent: `{
				"organization_name": "Test Org",
				"organization_bic": "TESTBIC1",
				"organization_iban": "TEST123456789",
				"credit_transfers": [
					{
						"amount": "1000.00",
						"counterparty_name": "John Doe",
						"counterparty_bic": "JOHNDOEBIC",
						"counterparty_iban": "JOHNDOE987654321",
						"description": "Test transfer"
					}
				]
			}`,
			setupMock: func(mockService *mock.TransferServiceMock) {
				mockService.EXPECT().
					BulkTransfer(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("insufficient funds"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Message: "insufficient funds",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock.NewTransferServiceMock(ctrl)
			tt.setupMock(mockService)

			// Create a new validator instance
			validate = validator.New()

			api := &apiDetails{
				service: mockService,
			}

			router := gin.New()
			router.POST("/transfers", api.BulkTransfer)

			// Create a multipart form with the file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "test.json")
			part.Write([]byte(tt.fileContent))
			writer.Close()

			req, _ := http.NewRequest("POST", "/transfers", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response interface{}
			if tt.expectedStatusCode == http.StatusCreated {
				var bulkTransferResponse BulkTransferResponse
				json.Unmarshal(w.Body.Bytes(), &bulkTransferResponse)
				response = bulkTransferResponse
			} else {
				var errorResponse ErrorResponse
				json.Unmarshal(w.Body.Bytes(), &errorResponse)
				response = errorResponse
			}

			assert.Equal(t, tt.expectedResponse, response)
		})
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		want    int64
		wantErr bool
	}{
		{"Whole number", "100", 10000, false},
		{"One decimal place", "100.5", 10050, false},
		{"Two decimal places", "100.55", 10055, false},
		{"Zero", "0", 0, false},
		{"Zero with decimal", "0.00", 0, false},
		{"Large number", "1000000", 100000000, false},
		{"Small fraction", "0.01", 1, false},
		{"Invalid format", "100.555", 0, true},
		{"Non-numeric", "abc", 0, true},
		{"Empty string", "", 0, true},
		{"Negative number", "-100", 0, true},
		{"Multiple decimal points", "100.55.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}
