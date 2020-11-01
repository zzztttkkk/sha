package main

import (
	"github.com/zzztttkkk/suna/rbac"
	"regexp"
	"time"

	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	_ "github.com/zzztttkkk/suna/sqls/drivers/postgres"
	"github.com/zzztttkkk/suna/validator"
)

var RedisUrl = "redis://127.0.0.1:6379"
var SqlUrl = "postgres://postgres:123456@localhost/suna_examples_simple?sslmode=disable"

type User struct {
	id int64
}

func (user *User) GetId() int64 {
	return user.id
}

func main() {
	conf := config.Default()
	conf.Redis.Mode = "singleton"
	conf.Redis.Nodes = append(conf.Redis.Nodes, RedisUrl)
	conf.Sql.Leader = SqlUrl
	conf.Sql.Logging = true

	conf.Done()

	suna.Init(
		&suna.InitOption{
			Config: conf,
			Authenticator: auth.AuthenticatorFunc(
				func(ctx *fasthttp.RequestCtx) (auth.User, bool) {
					if gotils.B2S(ctx.FormValue("token")) == "123456" {
						return &User{id: 1}, true
					}
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
			"${userId:06d} ${remote} ${method} ${path} "+
				"UserAgent:${reqHeader/User-Agent} "+
				"${statusCode} ${statusText} ${cost:03d}ms\n",
			nil,
		),
		middleware.NewRateLimiter(
			10,
			time.Second*1, 10,
			func(ctx *fasthttp.RequestCtx) string { return ctx.RemoteAddr().String() },
		),
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
	root.GET("/doc", root.DocHandler())

	loader := router.NewLoader()
	loader.AddChild("/rbac", rbac.Loader())

	loader.RunAsHttpServer(root, conf.Http.Address, conf.Http.TLS.Cert, conf.Http.TLS.Key)
}
