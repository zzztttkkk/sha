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

	m.HTTP("GET", "/", makeTestHandler(0))
	m.HTTP("GET", "/s/filename:*", makeTestHandler(2))
	m.HTTP("GET", "/simple/", makeTestHandler(3))
	m.HTTP("GET", "/simple/b", makeTestHandler(4))
	m.HTTP("GET", "/c/:d/:e", makeTestHandler(5))
	m.HTTP("GET", "/e/:f", makeTestHandler(6))
	m.HTTP("GET", "/g/:f/jjj", makeTestHandler(7))
	m.HTTP("POST", "/g/:f/jjj", makeTestHandler(8))
	m.HTTP("GET", "/fi/", makeTestHandler(1))
	m.HTTP("GET", "/虎虎虎/", makeTestHandler(56))
	m.HTTP("GET", "/qwer/:n/d/:m/sp:*", makeTestHandler(17))

	m.Print()

	s := Default(m)
	s.ListenAndServe()
}
