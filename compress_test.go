package sha

import (
	"fmt"
	"strings"
	"testing"
)

func TestRequestCtx_AutoCompress(t *testing.T) {
	s := Default(nil)
	s.option.MaxConnectionKeepAlive.Duration = 0

	mux := NewMux(nil, nil)
	s.Handler = mux

	mux.HTTP(
		"get",
		"/",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			fmt.Println(&ctx.Response.Header)
			_, _ = ctx.WriteString(strings.Repeat("Hello!", 100))
			fmt.Printf("%p %p %p\n", ctx, ctx.Response.compressWriter, ctx.Response.compressWriterPool)
			ctx.Close()
		}),
	)

	mux.HTTP(
		"get",
		"/a",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			_, _ = ctx.WriteString(strings.Repeat("Hello!", 100))
			ctx.Response.ResetBodyBuffer()
			_, _ = ctx.WriteString(strings.Repeat("World", 100))
		}),
	)

	s.ListenAndServe()
}
