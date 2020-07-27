package output

import (
	"fmt"
	"net/http"
)

type Err interface {
	Error() string
	StatusCode() int
	Message() *Message
}

type _Error struct {
	hcode  int
	errno  int
	errmsg string
	msg    *Message
}

func (e *_Error) Error() string {
	return fmt.Sprintf("%d %s", e.errno, e.errmsg)
}

func (e *_Error) StatusCode() int {
	return e.hcode
}

func (e *_Error) Message() *Message {
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

func NewError(httpCode, customErrno int, errmsg string) *_Error {
	return &_Error{
		hcode:  httpCode,
		errno:  customErrno,
		errmsg: errmsg,
		msg:    nil,
	}
}

var HttpErrors = map[int]*_Error{}

func init() {
	for v := 100; v < 550; v++ {
		txt := http.StatusText(v)
		if len(txt) < 1 {
			continue
		}
		HttpErrors[v] = NewError(v, -1, txt)
	}
}
