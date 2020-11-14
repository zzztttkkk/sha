package suna

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_WebSock(t *testing.T) {
	s := Server{
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			fmt.Println(string(ctx.Request.Path), &ctx.Request.Query)
			_, _ = ctx.WriteString("Hello World")
		},
	)

	go func() {
		time.Sleep(time.Second)
		s.Http1xProtocol.SubProtocols = map[string]Protocol{}
		s.Http1xProtocol.SubProtocols["websocket"] = &WebsocketProtocol{
			server:   &s,
			PreCheck: nil,
		}
	}()

	s.ListenAndServe()
}
