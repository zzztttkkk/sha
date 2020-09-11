package main

import (
	"regexp"

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
	"github.com/zzztttkkk/suna/validator"
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
	root.PanicHandler = output.RecoverAndLogging
	root.NotFound = output.NotFound
	root.MethodNotAllowed = output.MethodNotAllowed

	root.Use(
		middleware.NewAccessLogger(
			"{userId} {remote} {method} {path} UserAgent:{reqHeader/User-Agent} {statusCode} {statusText} {cost}ms\n",
			nil,
			nil,
		).AsMiddleware(),
	)

	var emptyRegexp = regexp.MustCompile(`\s+`)
	var emptyBytes = []byte("")

	validator.RegisterFunc(
		"username",
		func(data []byte) ([]byte, bool) {
			v := emptyRegexp.ReplaceAll(data, emptyBytes)
			return v, len(v) > 3
		},
		"remove all space characters and make sure the length is greater than 3",
	)

	type Form struct {
		Ignore string `validator:"-"`
		Name   string `validator:"L<3-20>;F<username>;D<null>;I<username>"`
	}

	root.GETWithDoc(
		"/hello",
		func(ctx *fasthttp.RequestCtx) {
			form := Form{}
			if !validator.Validate(ctx, &form) {
				return
			}
			output.MsgOK(ctx, form.Name)
		},
		validator.MakeDoc(Form{}, "print hello."),
	)
	root.BindDocHandler("/doc", nil)

	loader := utils.NewLoader()
	loader.AddChild("/rbac", rbac.Loader())

	loader.RunAsHttpServer(root, &conf)
}
