package output

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

type MessageErr interface {
	Error() string
	StatusCode() int
	Message() *Message
}

type msgError struct {
	hcode  int
	errno  int
	errmsg string
	msg    *Message
}

func (e *msgError) Error() string {
	return fmt.Sprintf("%d %s", e.errno, e.errmsg)
}

func (e *msgError) StatusCode() int {
	return e.hcode
}

func (e *msgError) Message() *Message {
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

//revive:disable-next-line
func NewError(httpCode, customErrno int, errmsg string) *msgError {
	return &msgError{
		hcode:  httpCode,
		errno:  customErrno,
		errmsg: errmsg,
		msg:    nil,
	}
}

var HttpErrors = map[int]*msgError{}

func init() {
	for v := 100; v < 600; v++ {
		txt := fasthttp.StatusMessage(v)
		if len(txt) < 1 {
			continue
		}
		HttpErrors[v] = NewError(v, -1, txt)
	}
}
