package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
)

func init() {
	loader.Http(
		func(router router.Router) {
			router.GET(
				"/status",
				newPermChecker(
					"rbac.status",
					func(ctx *fasthttp.RequestCtx) {
						g.RLock()
						defer g.RUnlock()
						output.MsgOK(ctx, output.M{"errors": errs})
					},
				),
			)
		},
	)
}
