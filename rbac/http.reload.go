package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/router"
)

// get: /reload
func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(R router.Router) {
					R.GETWithDoc(
						"/reload",
						newPAllPermChecker(
							"rbac.reload",
							func(ctx *fasthttp.RequestCtx) {
								Load(context.Background())
							},
						),
						router.NewDoc(""),
					)
				},
			)
		},
	)
}
