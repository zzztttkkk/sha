package output

import (
	"net/http"
)

type HttpError interface {
	Error() string
	StatusCode() int
	Message() *Message
}

type _HttpErrorT struct {
	hcode  int
	errno  int
	errmsg string
	msg    *Message
}

func (e *_HttpErrorT) Error() string {
	return e.errmsg
}

func (e *_HttpErrorT) StatusCode() int {
	return e.hcode
}

func (e *_HttpErrorT) Message() *Message {
	if e.msg != nil {
		return e.msg
	}
	e.msg = &Message{
		Errno:  e.errno,
		ErrMsg: e.errmsg,
		Data:   nil,
	}
	return e.msg
}

func NewHttpError(httpCode, customErrno int, errmsg string) *_HttpErrorT {
	return &_HttpErrorT{
		hcode:  httpCode,
		errno:  customErrno,
		errmsg: errmsg,
		msg:    nil,
	}
}

var StdErrors = map[int]*_HttpErrorT{}

func init() {
	for v := 100; v < 550; v++ {
		txt := http.StatusText(v)
		if len(txt) < 1 {
			continue
		}
		StdErrors[v] = NewHttpError(v, -1, txt)
	}
}
