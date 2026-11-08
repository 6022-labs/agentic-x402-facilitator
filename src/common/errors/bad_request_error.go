package errors

import "fmt"

// BadRequestError maps to HTTP 400 for a request transport could not decode (body, params, query).
type BadRequestError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewBadRequestError(code, message string) *BadRequestError {
	return &BadRequestError{Code: code, Message: message}
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
