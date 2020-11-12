package suna

import (
	"context"
	"fmt"
	"testing"
)

func TestServer_Run(t *testing.T) {
	s := Server{
		http1xProtocol: &Http1xProtocol{
			MaxFirstLintSize: 2048,
			MaxHeadersSize:   4096,
			MaxBodySize:      4096 * 1024,
			ReadTimeout:      0,
			WriteTimeout:     0,
			Version:          []byte("HTTP/1.1"),
			ReadBufferSize:   128,
		},
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			fmt.Println(&ctx.Request.Header)
			_, _ = ctx.WriteString("Hello World")
		},
	)

	s.ListenAndServe()
}
