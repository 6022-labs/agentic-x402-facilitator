package errors

import "fmt"

// UnavailableError maps to HTTP 502; Message is fixed caller text, the dependency's own error stays in Cause.
type UnavailableError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func NewUnavailableError(code, message string, cause error) *UnavailableError {
	return &UnavailableError{Code: code, Message: message, Cause: cause}
}

func (e *UnavailableError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *UnavailableError) Unwrap() error {
	return e.Cause
}
