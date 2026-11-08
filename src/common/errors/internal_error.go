package errors

import "fmt"

// InternalErrorMessage is the only message a 500 body carries; the cause can leak RPC URLs with API keys.
const InternalErrorMessage = "internal server error"

// InternalError maps to HTTP 500.
type InternalError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func NewInternalError(code string, cause error) *InternalError {
	return &InternalError{Code: code, Message: InternalErrorMessage, Cause: cause}
}

func (e *InternalError) Error() string {
	if e.Cause == nil {
		return e.Code
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Cause)
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}
