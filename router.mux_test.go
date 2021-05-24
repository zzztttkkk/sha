package sha

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func makeHandler(v int) RequestHandler {
	return RequestHandlerFunc(func(ctx *RequestCtx) {
		fmt.Println(v, &ctx.Request.URL.Params)
		_ = ctx.WriteString("Hello!")
	})
}

func TestMux(t *testing.T) {
	mux := NewMux(nil)

	mux.Use(
		MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			_ = ctx.WriteString("global middleware 1\n")
			next()
		}),
		MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			_= ctx.WriteString("global middleware 2\n")
			next()
		}),
	)

	mux.HTTP("get", "/", makeHandler(3))
	mux.HTTP("post", "/", makeHandler(31))

	mux.HTTP("get", "/book/{name}/", makeHandler(4))
	mux.HTTP("get", "/book/{name}/{chapter}/", makeHandler(5))
	mux.HTTPWithOptions(
		&HandlerOptions{
			Middlewares: []Middleware{
				MiddlewareFunc(func(ctx *RequestCtx, next func()) {
					_ = ctx.WriteString("handler middleware 1\n")
					next()
				}),
			},
		},
		"get", "/foo", makeHandler(51),
	)

	mux.FileSystem(nil, "get", "/os/src/{filepath:*}", http.Dir("./"), true)
	mux.File(nil, "get", "/LICENSE.txt", "./LICENSE")
	mux.HTTP("get", "/embed/src/{filepath:*}", NewEmbedFSHandler(&ef, time.Time{}, nil))

	groupA := mux.NewGroup("/a")
	groupA.Use(
		MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			_ = ctx.WriteString("A middleware 1\n")
			next()
		}),
		MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			_ = ctx.WriteString("A middleware 2\n")
			next()
		}),
	)

	groupA.HTTP("get", "/", makeHandler(6))
	groupA.HTTP("get", "/files/{filename:*}", makeHandler(7))

	fmt.Print(mux)
	ListenAndServe("", mux)
}
