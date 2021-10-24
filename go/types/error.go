package types

import "fmt"

const (
	ErrCodeBadRequest   = "bad_request"
	ErrCodeUnknown      = "unknown"
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeNotFound     = "not_found"
	ErrCodeInternal     = "internal_error"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (ae Error) Error() string {
	return fmt.Sprintf("%s: %s", ae.Code, ae.Message)
}

func (ae Error) String() string {
	return ae.Error()
}

func (ae Error) Unwrap() error {
	return fmt.Errorf("%s", ae.Error())
}

func NewError(code string, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}
