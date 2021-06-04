package validator

import (
	"fmt"
	"net/http"
)

const (
	MissingRequired = _FormErrorType(iota)
	BadValue
)

type Error struct {
	FormName string
	Type     _FormErrorType
	Wrapped  error
}

var CustomError func(fe *Error) string

func init() {
	CustomError = func(fe *Error) string {
		if fe.Wrapped == nil {
			return fmt.Sprintf("ValidateError: %s; field `%s`", fe.Type, fe.FormName)
		}
		return fmt.Sprintf("ValidateError: %s, %s; field `%s`", fe.Type, fe.Wrapped.Error(), fe.FormName)
	}
}

func (e *Error) Error() string {
	return CustomError(e)
}

func (e *Error) StatusCode() int {
	return http.StatusBadRequest
}
