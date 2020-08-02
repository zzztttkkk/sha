package suna

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/utils"
	"testing"
)

func TestInit(t *testing.T) {
	Init(nil)

	root := router.New()
	loader := utils.NewLoader()
	root.GET(
		"/",
		func(ctx *fasthttp.RequestCtx) {
			fmt.Fprintln(ctx, "Hello World!")
		},
	)
	loader.RunAsHttpServer(root, ":8080")
}
