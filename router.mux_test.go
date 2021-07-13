package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"net/http"
	"testing"
	"time"
)

func makeHandler(v int) RequestCtxHandler {
	return RequestCtxHandlerFunc(func(ctx *RequestCtx) {
		fmt.Println(v, &ctx.Request.URL.Params)
		_ = ctx.WriteString("Hello!")
	})
}

type RandIDGenerator struct{}

func (RandIDGenerator) Size() int { return 16 }

func (RandIDGenerator) Generate(v []byte) {
	for i := 0; i < 16; i++ {
		v[i] = utils.RandByte(nil)
	}
}

func init() {
	UniqueIDGenerator = RandIDGenerator{}
}

func TestMux(t *testing.T) {
	mux := NewMux(nil)

	mux.Use(
		MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			fmt.Printf("global middleware 1: %p %s\r\n", ctx, ctx.Request.GUID())
			next()
		}),
		MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			fmt.Println("global middleware 2")
			next()
		}),
	)

	mux.HTTP("get", "/", makeHandler(3))
	mux.HTTP("post", "/", makeHandler(31))

	mux.HTTP("get", "/book/{name}/", makeHandler(4))
	mux.HTTP("get", "/book/{name}/{chapter}/", makeHandler(5))
	mux.HTTPWithOptions(
		&RouteOptions{
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
