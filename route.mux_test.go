package sha

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
	m.AutoSlashRedirect = true

	//rawHandlerMap.REST("GET", "/current:*", h)
	m.REST("GET", "/", makeTestHandler(0))
	//rawHandlerMap.REST("GET", "/fi", makeTestHandler(1))
	m.REST("GET", "/s/filename:*", makeTestHandler(2))
	m.REST("GET", "/simple/", makeTestHandler(3))
	m.REST("GET", "/simple/b", makeTestHandler(4))
	m.REST("GET", "/c/:d/:e", makeTestHandler(5))
	m.REST("GET", "/e/:f", makeTestHandler(6))
	m.REST("GET", "/g/:f/jjj", makeTestHandler(7))
	m.REST("POST", "/g/:f/jjj", makeTestHandler(8))
	m.REST("GET", "/fi/", makeTestHandler(1))
	m.REST("GET", "/虎虎虎/", makeTestHandler(1))

	m.Print(false, false)

	s := Default(m)
	s.ListenAndServe()
}
