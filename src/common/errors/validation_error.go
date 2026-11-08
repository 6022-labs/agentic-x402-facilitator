package errors

import "fmt"

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}
