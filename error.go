package suna

import (
	"fmt"
	"net/http"
)

type HttpError interface {
	Error() string
	StatusCode() int
	Message() []byte
	Header() *Header
}

type _StdError struct {
	status  int
	message []byte
	header  Header
}

func (err *_StdError) Error() string {
	return fmt.Sprintf("%d %s", err.status, err.message)
}

func (err *_StdError) StatusCode() int {
	return err.status
}

func (err *_StdError) Message() []byte {
	return err.message
}

func (err *_StdError) Header() *Header {
	return &err.header
}

var StdHttpErrors = map[int]HttpError{}

func newStdError(msg string) HttpError {
	return &_StdError{
		status:  http.StatusInternalServerError,
		message: []byte(msg),
	}
}

var (
	ErrBadConnection               HttpError
	ErrRequestUrlTooLong           HttpError
	ErrRequestHeaderFieldsTooLarge HttpError
	ErrRequestEntityTooLarge       HttpError
)

func init() {
	for i := 400; i < 600; i++ {
		txt := http.StatusText(i)
		if len(txt) < 1 {
			continue
		}
		StdHttpErrors[i] = &_StdError{
			status:  i,
			message: []byte(txt),
		}
	}

	ErrBadConnection = StdHttpErrors[http.StatusBadRequest]
	ErrRequestUrlTooLong = StdHttpErrors[http.StatusRequestURITooLong]
	ErrRequestHeaderFieldsTooLarge = StdHttpErrors[http.StatusRequestHeaderFieldsTooLarge]
	ErrRequestEntityTooLarge = StdHttpErrors[http.StatusRequestEntityTooLarge]
}
