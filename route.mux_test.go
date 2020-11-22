package suna

import (
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
	m := NewMux("", nil)
	m.AutoOptions = true
	m.AutoRedirect = true

	//m.AddHandler("GET", "/current:*", h)
	m.AddHandler("GET", "/", makeTestHandler(0))
	//m.AddHandler("GET", "/fi", makeTestHandler(1))
	m.AddHandler("GET", "/s/filename:*", makeTestHandler(2))
	m.AddHandler("GET", "/simple/", makeTestHandler(3))
	m.AddHandler("GET", "/simple/b", makeTestHandler(4))
	m.AddHandler("GET", "/c/:d/:e", makeTestHandler(5))
	m.AddHandler("GET", "/e/:f", makeTestHandler(6))
	m.AddHandler("GET", "/g/:f/jjj", makeTestHandler(7))
	m.AddHandler("POST", "/g/:f/jjj", makeTestHandler(8))
	m.AddHandler("GET", "/fi/", makeTestHandler(1))

	s := Default()

	s.Handler = m

	s.ListenAndServe()
}
