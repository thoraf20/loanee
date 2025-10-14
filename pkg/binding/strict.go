package binding

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func StrictBindJSON[T any](c *gin.Context) (*T, error) {
	var obj T

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&obj); 
	err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); 
	ok {
		if err := v.Struct(obj); err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
	}

	return &obj, nil
}
