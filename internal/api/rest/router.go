package rest

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	swagFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var validate *validator.Validate

type ErrorResponse struct {
	Message string `json:"message"`
}

func createErrorResponse(c *gin.Context, code int, message string) {
	c.IndentedJSON(code, &ErrorResponse{
		Message: message,
	})
}

func (api *apiDetails) setupRouter() *gin.Engine {
	validate = validator.New()
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowHeaders = append(config.AllowHeaders, "Access-Control-Allow-Origin")
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	apiV1 := r.Group("/api/v1")
	apiV1.GET("/swagger/*any", ginSwagger.WrapHandler(swagFiles.Handler))
	apiV1.GET("/health", api.health)
	apiV1.POST("/transfers", api.BulkTransfer)
	return r
}
