package router

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"testing"
)

func TestNewRouter(t *testing.T) {
	ctx := fasthttp.RequestCtx{Request: *fasthttp.AcquireRequest()}
	a, ok := ctx.UserValue("s").(int)
	fmt.Println(a, ok, ctx.UserValue("s"))
}

func Test_RouterT_GET(t *testing.T) {
	root := New()
	a := root.SubGroup("a")
	a.GET("/", func(ctx *fasthttp.RequestCtx) {})

	aa := a.SubGroup("aa")
	aa.GET("/xx", func(ctx *fasthttp.RequestCtx) {})

	aaa := aa.SubGroup("aaa")
	aaa.GET("/ccc", func(ctx *fasthttp.RequestCtx) {})

	b := root.SubGroup("b")
	b.GET("/index", func(ctx *fasthttp.RequestCtx) {})

	fmt.Println(root.List(), root.path)
	fmt.Println(b.List(), b.path)
	fmt.Println(a.List(), a.path, aa.path, aaa.path)
}
