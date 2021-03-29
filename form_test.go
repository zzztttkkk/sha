package sha

import (
	"fmt"
	"testing"
)

func TestFormFiles(t *testing.T) {
	ListenAndServe("", RequestHandlerFunc(func(ctx *RequestCtx) {
		fmt.Println(ctx.Request.Files(), ctx.Request.BodyForm())
		fmt.Println(111)
	}))
}
