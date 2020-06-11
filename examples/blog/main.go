package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend"
	bctxs "github.com/zzztttkkk/snow/examples/blog/backend/ctxs"
	"github.com/zzztttkkk/snow/examples/blog/backend/services"
	"github.com/zzztttkkk/snow/mware"
	sctxs "github.com/zzztttkkk/snow/mware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
	"os"
	"time"
)

func main() {
	conf := &snow.Config{}
	conf.IniFiles = append(conf.IniFiles, os.Getenv("ProjectRoot")+"/examples/blog/conf.ini")
	conf.UserReader = bctxs.GetUid
	snow.Init(conf)

	backend.Init()

	root := router.New()
	root.Use(
		mware.NewRateLimiter(sctxs.GetRemoteIpHash, time.Second, 30),
		mware.SessionHandler,
	)

	root.PanicHandler = output.Recover
	root.NotFound = func(ctx *fasthttp.RequestCtx) { output.StdError(ctx, fasthttp.StatusNotFound) }
	root.MethodNotAllowed = func(ctx *fasthttp.RequestCtx) { output.StdError(ctx, fasthttp.StatusMethodNotAllowed) }

	snow.RunAsHttpServer(services.Loader, root)
}
