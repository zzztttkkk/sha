package suna

import (
	"context"
	"fmt"
	"testing"
)

func TestServer_Run(t *testing.T) {
	s := Server{
		HttpProtocol: HttpProtocol{
			MaxFirstLintSize: 2048,
			MaxHeadersSize:   4096,
			MaxBodySize:      4096 * 1024,
			ReadTimeout:      0,
			WriteTimeout:     0,
		},
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			ctx.UseResponseBuffer(false)
			fmt.Println(&ctx.Request.Header)
			_, _ = ctx.Write(
				[]byte("HTTP/1.1 200 OK\r\nContent-Length: 7\r\nServer: suna\r\nConnection: keep-alive\r\n\r\nSpring!"),
			)
		},
	)

	s.Run()
}
