package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/output"
)

func init() {
	loader.Http(
		func(router router.Router) {
			router.GET(
				"/status/errors",
				newPermChecker(
					"rbac.status.errors",
					func(ctx *fasthttp.RequestCtx) {
						g.RLock()
						defer g.RUnlock()
						output.MsgOK(ctx, errs)
					},
				),
			)
		},
	)
}
