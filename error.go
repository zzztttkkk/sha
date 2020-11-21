package suna

import (
	"fmt"
	"net/http"
)

type HttpError interface {
	Error() string
	StatusCode() int
}

type HttpResponseError interface {
	HttpError
	Header() *Header
	Body() []byte
}

type StatusError int

func (err StatusError) Error() string {
	return fmt.Sprintf("%d %s", err, http.StatusText(int(err)))
}

func (err StatusError) StatusCode() int {
	return int(err)
}

func (err StatusError) Header() *Header {
	return nil
}

func (err StatusError) Body() []byte {
	i := int(err)
	if i < 0 || i > 599 {
		panic(ErrUnknownResponseStatusCode)
	}
	ret := statusTextMap[i]
	if len(ret) < 1 {
		panic(ErrUnknownResponseStatusCode)
	}
	return ret
}

var (
	ErrBadConnection               = StatusError(http.StatusBadRequest)
	ErrRequestUrlTooLong           = StatusError(http.StatusRequestURITooLong)
	ErrRequestHeaderFieldsTooLarge = StatusError(http.StatusRequestHeaderFieldsTooLarge)
	ErrRequestEntityTooLarge       = StatusError(http.StatusRequestEntityTooLarge)
)
