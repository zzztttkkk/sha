package suna

import (
	"strings"
	"testing"
)

func TestRequestCtx_AutoCompress(t *testing.T) {
	s := Default(nil)

	mux := NewMux("", nil)
	s.Handler = mux

	mux.REST(
		"get",
		"/",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			_, _ = ctx.WriteString(strings.Repeat("Hello!", 100))
		}),
	)

	mux.REST(
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
