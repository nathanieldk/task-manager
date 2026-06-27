package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		var messages []string
		for _, e := range err.(validator.ValidationErrors) {
			messages = append(messages, formatValidationError(e))
		}
		return echo.NewHTTPError(422, strings.Join(messages, "; "))
	}
	return nil
}

func formatValidationError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
