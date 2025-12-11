package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) Validate(data interface{}) error {
	if err := v.validate.Struct(data); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return formatValidationErrors(validationErrors)
		}
		return err
	}
	return nil
}

func formatValidationErrors(errs validator.ValidationErrors) error {
	var messages []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", strings.ToLower(err.Field())))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must not exceed %s characters", strings.ToLower(err.Field()), err.Param()))
		case "gte":
			messages = append(messages, fmt.Sprintf("%s must be greater than or equal to %s", strings.ToLower(err.Field()), err.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", strings.ToLower(err.Field())))
		}
	}
	return fmt.Errorf(strings.Join(messages, ", "))
}
