package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/internal"
	"net/http"
)

type HttpError interface {
	Error() string
	StatusCode() int
}

type HttpResponseError interface {
	HttpError
	Header(*Header)
	Body() []byte
}

type StatusError int

func (err StatusError) Error() string {
	return fmt.Sprintf("%d %s", err, http.StatusText(int(err)))
}

func (err StatusError) StatusCode() int { return int(err) }

func (err StatusError) Header(_ *Header) {}

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

type _RedirectError struct {
	status int
	uri    string
}

func (err *_RedirectError) Error() string { return "" }

func (err *_RedirectError) StatusCode() int { return err.status }

func (err *_RedirectError) Header(h *Header) { h.Set(HeaderLocation, internal.B(err.uri)) }

func (err *_RedirectError) Body() []byte { return nil }

func RedirectPermanently(uri string) { panic(&_RedirectError{status: StatusMovedPermanently, uri: uri}) }

func RedirectTemporarily(uri string) { panic(&_RedirectError{status: StatusFound, uri: uri}) }
