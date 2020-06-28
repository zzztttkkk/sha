package main

import (
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/examples/blog/backend/models"

	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend"
	"github.com/zzztttkkk/snow/examples/blog/backend/services"
	"github.com/zzztttkkk/snow/middleware"
	"github.com/zzztttkkk/snow/middleware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
)

func main() {
	snow.Init(
		&snow.InitOption{
			IniFiles:    []string{os.Getenv("PROJECT_ROOT") + "/examples/blog/conf.ini"},
			UserFetcher: models.UserOperator.GetById,
		},
	)

	backend.Init()

	root := router.New()
	root.Use(
		middleware.NewRateLimitMiddleware(ctxs.RemoteIpHash, time.Second, 30),
		middleware.SessionAndAuthMiddleware,
	)

	root.PanicHandler = output.Recover
	root.NotFound = func(ctx *fasthttp.RequestCtx) { output.StdError(ctx, fasthttp.StatusNotFound) }
	root.MethodNotAllowed = func(ctx *fasthttp.RequestCtx) { output.StdError(ctx, fasthttp.StatusMethodNotAllowed) }

	snow.RunAsHttpServer(services.Loader, root)
}
