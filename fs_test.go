package suna

import (
	"net/http"
	"testing"
)

func TestFs(t *testing.T) {
	server := Default(nil)
	mux := NewMux("", nil)
	mux.StaticFile(
		"get",
		"/suna/filename:*",
		http.Dir("./"),
		true,
		MiddlewareFunc(
			func(ctx *RequestCtx, next func()) {
				ctx.AutoCompress()
				next()
			},
		),
	)

	server.Handler = mux
	server.ListenAndServe()
}
