package category

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/middleware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
)

var Loader = snow.NewLoader()

func init() {
	Loader.Http(
		func(router *router.Router) {
			categoryCache := middleware.NewRedCache(&middleware.RedCacheOption{ExpireSeconds: 1800})

			router.GET(
				"/all",
				categoryCache.AsHandler(
					func(ctx *fasthttp.RequestCtx) {
						output.MsgOK(ctx, output.M{"lst": models.CategoryOperator.List(ctx)})
					},
				),
			)
		},
	)
}
