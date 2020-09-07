package main

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/utils"
)

var RedisUrl = "redis://127.0.0.1:6379"
var SqlUrl = ":memory:"
var SqlDriver = "sqlite3"

func main() {
	conf := config.Default()
	conf.Redis.Mode = "singleton"
	conf.Redis.Nodes = append(conf.Redis.Nodes, RedisUrl)
	conf.Sql.Leader = SqlUrl
	conf.Sql.Driver = SqlDriver
	conf.Sql.Logging = true

	conf.Done()

	suna.Init(
		&suna.InitOption{
			Config: &conf,
			Authenticator: auth.AuthenticatorFunc(
				func(ctx *fasthttp.RequestCtx) (auth.User, bool) {
					return nil, false
				},
			),
		},
	)

	root := router.New(nil)

	root.PanicHandler = output.Recover

	root.Use(
		middleware.NewAccessLogger(
			"UserId:{userId} <{method} {path}> UserAgent:{reqHeader/User-Agent} <{statusCode} {statusText}> {cost}ms {errStack}",
			nil,
			&middleware.AccessLoggingOption{
				Enabled:      true,
				DurationUnit: time.Millisecond,
			},
		).AsMiddleware(),
	)

	root.GET(
		"/hello",
		func(ctx *fasthttp.RequestCtx) {
			output.MsgOK(ctx, "World!")
		},
	)

	loader := utils.NewLoader()
	loader.AddChild("/rbac", rbac.Loader())

	loader.RunAsHttpServer(root, &conf)
}
