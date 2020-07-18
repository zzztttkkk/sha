package services

import (
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/account"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/category"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/debug"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/post"
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/middleware/ctxs"
	"github.com/zzztttkkk/snow/router"
)

var Loader = snow.NewLoader()

func init() {
	Loader.AddChild("account", account.Loader)
	Loader.AddChild("category", category.Loader)
	Loader.AddChild("post", post.Loader)

	Loader.Http(
		func(router *router.Router) {
			router.GET(
				"/captcha.png",
				func(ctx *fasthttp.RequestCtx) {
					ctxs.Session(ctx).CaptchaGenerate(ctx)
				},
			)
		},
	)

	internal.LazyExecutor.Register(
		func(kwargs snow.Kwargs) {
			conf := kwargs["config"].(*ini.Config)

			if conf.IsDebug() {
				Loader.AddChild("debug", debug.Loader)
			}
		},
	)
}
