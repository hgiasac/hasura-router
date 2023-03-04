package types

import "fmt"

const (
	ErrCodeBadRequest   = "bad_request"
	ErrCodeUnknown      = "unknown"
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeNotFound     = "not_found"
	ErrCodeInternal     = "internal_error"
)

// Error represents the action error response object.
type Error struct {
	Code       string                 `json:"code,omitempty"`
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Error implements the error interface.
func (ae Error) Error() string {
	msg := fmt.Sprintf("%s; extensions: %+v", ae.Message, ae.Extensions)
	if ae.Code == "" {
		return msg
	}

	return fmt.Sprintf("%s: %s", ae.Code, msg)
}

// String handle the error interface.
func (ae Error) String() string {
	return ae.Error()
}

// Unwrap unwrap the error.
func (ae Error) Unwrap() error {
	return fmt.Errorf("%s", ae.Error())
}

// NewError creates an Error instance
func NewError(code string, message string) Error {
	extensions := make(map[string]interface{})
	if code != "" {
		extensions["code"] = code
	}
	return Error{
		Code:       code,
		Message:    message,
		Extensions: extensions,
	}
}
