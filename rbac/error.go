package rbac

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/output"
)

type Error struct {
	SubjectId      int64
	PermissionName string
}

func (err *Error) StatusCode() int {
	return fasthttp.StatusForbidden
}

func (err *Error) Error() string {
	return fmt.Sprintf("snow.rbac: `%d` has no permission `%s`", err.SubjectId, err.PermissionName)
}

func (err *Error) Message() *output.Message {
	return &output.Message{
		Errno:  -1,
		ErrMsg: "permission denied",
	}
}
