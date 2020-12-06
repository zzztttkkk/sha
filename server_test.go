package suna

import (
	"fmt"
	"strings"
	"testing"
)

func TestServer_Run(t *testing.T) {
	server := Default(nil)

	server.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			ctx.AutoCompress()
			_, _ = ctx.WriteString(strings.Repeat("Hello World", 100))
			fmt.Printf("%s %p\n", ctx.Request.Path, ctx.conn)
		},
	)
	server.ListenAndServe()
}
