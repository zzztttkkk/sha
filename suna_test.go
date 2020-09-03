package suna

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/utils"
	"testing"
	"time"
)

var REDIS_URL = "redis://127.0.0.1:6379"
var SQL_URL = ":memory:"
var SQL_DRIVER = "sqlite3"

func TestInit(t *testing.T) {
	conf := config.Default()
	conf.Redis.Mode = "singleton"
	conf.Redis.Nodes = append(conf.Redis.Nodes, REDIS_URL)
	conf.Sql.Leader = SQL_URL
	conf.Sql.Driver = SQL_DRIVER
	conf.Sql.Logging = true

	conf.Done()

	Init(
		&InitOption{
			Config: &conf,
			Authenticator: auth.AuthenticatorFunc(
				func(ctx *fasthttp.RequestCtx) (auth.User, bool) {
					return nil, false
				},
			),
		},
	)

	root := router.New()

	root.PanicHandler = output.Recover

	root.Use(
		middleware.NewAccessLogger(
			"{UserId} {Method} {Path} {StatusCode} {StatusText} {Cost}ms",
			nil,
			&middleware.AccessLoggingOption{
				Enabled:      true,
				DurationUnit: time.Millisecond,
			},
		).AsMiddleware(),
	)

	cors := middleware.NewCors(&middleware.CorsOption{AllowOrigins: "*"})
	root.Use(cors.AsMiddleware())
	root.GlobalOPTIONS = cors.AsOptionsHandler()

	root.GET(
		"/hello",
		func(ctx *fasthttp.RequestCtx) { output.MsgOK(ctx, "World!") },
	)

	loader := utils.NewLoader()
	loader.AddChild("/rbac", rbac.Loader())

	loader.RunAsHttpServer(root, &conf)
}
