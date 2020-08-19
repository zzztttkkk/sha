package main

import (
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna"
	"github.com/zzztttkkk/suna.example/config"
	"github.com/zzztttkkk/suna.example/internal"
	"github.com/zzztttkkk/suna.example/model"
	"github.com/zzztttkkk/suna.example/service"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/session"
	"github.com/zzztttkkk/suna/utils/toml"
	"log"
	"path/filepath"
	"time"
)

var configPath = flag.String("c", "", "contains some toml files")

func main() {
	suna.DisableReservedKeysWarning()
	router.DisableReservedKeysWarning()

	// load config from toml files, sort by filename.
	flag.Parse()
	if len(*configPath) < 1 {
		log.Fatalf("empty config path.(-c)")
	}
	files, err := filepath.Glob(*configPath + "/*.toml")
	if err != nil {
		log.Fatal(err)
	}
	if len(files) < 1 {
		log.Fatalf("empty conf path: `%s`", *configPath)
	}

	var example config.Example
	toml.FromFiles(&example, config.Default(), files...)

	// init suna
	suna.Init(
		&suna.InitOption{
			Config:        &example.Suna,
			Authenticator: model.UserOperator,
		},
	)

	// di
	internal.Dig.Provide(func() *config.Example { return &example })
	internal.Dig.Index(model.User{}) // trigger the init function.
	internal.Dig.Invoke()

	internal.LazyExecutor.Execute(nil)

	// router
	root := router.New()
	root.Use(
		middleware.NewAccessLogger(
			"{ReqMethod} {ReqPath} {UserId} {Cost}ms {ResStatusCode} {ResStatusText} {ErrStack}",
			nil,
			&middleware.AccessLoggingOption{
				Enabled:      !example.IsRelease(),
				DurationUnit: time.Millisecond,
			},
		).AsMiddleware(),

		middleware.NewRateLimiter(
			&middleware.RateLimiterOption{
				GetKey: func(ctx *fasthttp.RequestCtx) string { return "global:" + ctx.RemoteIP().String() },
				Rate:   30,
				Period: time.Second,
				Burst:  10,
			},
		).AsMiddleware(),
	)
	root.PanicHandler = output.Recover
	root.NotFound = output.NotFound
	root.MethodNotAllowed = output.MethodNotAllowed

	root.BindDocHandler("/doc")

	root.GET(
		"/captcha.png",
		func(ctx *fasthttp.RequestCtx) { session.New(ctx).CaptchaGenerateImage(ctx) },
	)
	root.GET(
		"/captcha.wav",
		func(ctx *fasthttp.RequestCtx) { session.New(ctx).CaptchaGenerateAudio(ctx) },
	)

	service.Loader.AddChild("/rbac", rbac.Loader())
	service.Loader.RunAsHttpServer(root, ":8080")
}
