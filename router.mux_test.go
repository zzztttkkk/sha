package suna

import (
	"context"
	"fmt"
	"testing"
)

func makeTestHandler(id int) RequestHandler {
	return RequestHandlerFunc(
		func(ctx *RequestCtx) {
			fmt.Println(id, string(ctx.Request.Path), &ctx.Request.Params)
		},
	)
}

func Test_Mux_AddHandler(t *testing.T) {
	m := newMux("")

	//m.AddHandler("GET", "/file:*", h)
	m.AddHandler("GET", "/", makeTestHandler(0))
	m.AddHandler("GET", "/fi", makeTestHandler(1))
	m.AddHandler("GET", "/s/filename:*", makeTestHandler(2))
	m.AddHandler("GET", "/a/", makeTestHandler(3))
	m.AddHandler("GET", "/a/b", makeTestHandler(4))
	m.AddHandler("GET", "/c/:d/:e", makeTestHandler(5))
	m.AddHandler("GET", "/e/:f", makeTestHandler(6))
	m.AddHandler("GET", "/g/:f/jjj", makeTestHandler(7))

	s := Server{
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	s.Handler = m

	s.ListenAndServe()
}
