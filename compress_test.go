package sha

import (
	"fmt"
	"strings"
	"testing"
)

func TestRequestCtx_AutoCompress(t *testing.T) {
	mux := NewMux(nil)

	mux.HTTP(
		"get",
		"/",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			fmt.Println(ctx.Response.Header())
			_ = ctx.WriteString(strings.Repeat("PgUP", 1000))
			fmt.Printf("%p %p %p\n", ctx, ctx.Response.cw, ctx.Response.cwPool)
			ctx.Close()
		}),
	)

	mux.HTTP(
		"get",
		"/a",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			_ = ctx.WriteString(strings.Repeat("Hello!", 100))
			ctx.Response.ResetBody()
			_ = ctx.WriteString(strings.Repeat("World", 100))
		}),
	)

	ListenAndServe("", mux)
}
