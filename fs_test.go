package sha

import (
	"net/http"
	"testing"
)

func TestFs(t *testing.T) {
	server := Default(nil)
	mux := NewMux("", nil)

	mux.FileSystem(
		http.Dir("./"),
		"get",
		"/sha/filename:*",
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
