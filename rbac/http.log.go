package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
)

func init() {
	type Form struct {
		Begin     int64
		End       int64
		Names     []string
		Operators []int64
	}

	loader.Http(
		func(router router.Router) {
			router.GET(
				"/log/list",
				newPermChecker(
					"admin.rbac.log.read",
					func(ctx *fasthttp.RequestCtx) {

					},
				),
			)
		},
	)
}
