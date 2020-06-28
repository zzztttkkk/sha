package debug

import (
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/secret"
)

var Loader = snow.NewLoader()

func init() {
	Loader.Http(
		func(router *router.Router) {
			router.GET(
				"/pwd/{name:*}",
				func(ctx *fasthttp.RequestCtx) {
					name, ok := ctx.UserValue("name").(string)
					if !ok {
						output.StdError(ctx, fasthttp.StatusBadRequest)
						return
					}
					output.MsgOK(ctx, string(secret.Default.Calc([]byte(name))))
				},
			)
		},
	)
}
