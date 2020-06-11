package category

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
)

var Loader = snow.NewLoader()

func init() {
	Loader.Http(
		func(router *router.Router) {
			router.Use(
				mware.NewRedCache(&mware.RedCacheOption{ExpireSeconds: 1800}).Handler,
			)

			router.GET(
				"/all",
				func(ctx *fasthttp.RequestCtx) {
					output.MsgOk(ctx, output.M{"lst": models.CategoryOperator.List(ctx)})
				},
			)
		},
	)
}
