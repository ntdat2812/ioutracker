package util

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func Validate(request any) []string {
	validate := validator.New()
	var validationErrors []string

	if err := validate.Struct(request); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			tag := err.Tag()
			validationErrors = append(validationErrors, fmt.Sprintf("Field %s is invalid: %s", field, tag))
		}
	}

	return validationErrors
}
