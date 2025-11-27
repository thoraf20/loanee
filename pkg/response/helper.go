package response

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	utils "github.com/thoraf20/loanee/internal/utils"
)

// HandleValidationError converts validator errors to a readable format
func HandleValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = field + " must be a valid email"
			case "min":
				errors[field] = field + " must be at least " + e.Param() + " characters"
			case "max":
				errors[field] = field + " must be at most " + e.Param() + " characters"
			case "gte":
				errors[field] = field + " must be greater than or equal to " + e.Param()
			case "lte":
				errors[field] = field + " must be less than or equal to " + e.Param()
			case "oneof":
				errors[field] = field + " must be one of: " + e.Param()
			default:
				errors[field] = field + " is invalid"
			}
		}
		
		utils.ValidationError(c, errors)
		return
	}

	utils.BadRequest(c, "Invalid request", err.Error())
}

// SuccessWithToken sends a success response with authentication token
func SuccessWithToken(c *gin.Context, message string, data interface{}, token string) {
	response := map[string]interface{}{
		"user":  data,
		"token": token,
	}
	utils.Success(c, 200, message, response)
}

// SuccessWithTokens sends a success response with access and refresh tokens
func SuccessWithTokens(c *gin.Context, message string, data interface{}, accessToken, refreshToken string) {
	response := map[string]interface{}{
		"user":          data,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	utils.Success(c, 200, message, response)
}