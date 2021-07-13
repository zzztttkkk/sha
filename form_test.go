package sha

import (
	"fmt"
	"testing"
)

func TestFormFiles(t *testing.T) {
	ListenAndServe("", RequestCtxHandlerFunc(func(ctx *RequestCtx) {
		fmt.Println(ctx.Request.Files(), ctx.Request.BodyForm())
		fmt.Println(111)
	}))
}
