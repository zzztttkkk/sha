package services

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/account"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/category"
	"github.com/zzztttkkk/snow/mware/ctxs"
	"github.com/zzztttkkk/snow/router"
)

var Loader = snow.NewLoader()

func init() {
	Loader.AddChild("account", account.Loader)
	Loader.AddChild("category", category.Loader)

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
}
