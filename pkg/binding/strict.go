package binding

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Errors  []ValidationError  `json:"errors,omitempty"`
}

func StrictBindJSON[T any](r *http.Request) (*T, *ValidationErrorResponse) {
	var obj T

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&obj); err != nil {
		return nil, &ValidationErrorResponse{
			Message: fmt.Sprintf("Invalid request body: %v", err.Error()),
		}
	}

	if err := validate.Struct(obj); err != nil {
		var validationErrors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Message: fmt.Sprintf("Field validation for '%s' failed on the '%s' rule", err.Field(), err.Tag()),
			})
		}

		return nil, &ValidationErrorResponse{
			Message: "Validation failed",
			Errors:  validationErrors,
		}
	}

	return &obj, nil
}
