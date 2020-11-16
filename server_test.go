package suna

import (
	"context"
	"strings"
	"testing"
)

func TestServer_Run(t *testing.T) {
	s := Server{
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			ctx.AutoCompress()
			ctx.Response.Header.SetContentType(MIMEText)
			_, _ = ctx.WriteString(strings.Repeat("Hello World", 100))
		},
	)
	s.ListenAndServe()
}
