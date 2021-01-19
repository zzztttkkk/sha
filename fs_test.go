package sha

import (
	"testing"
)

func TestFs(t *testing.T) {
	server := Default(nil)
	mux := NewMux("", nil)

	mux.FilePath(
		"./",
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
