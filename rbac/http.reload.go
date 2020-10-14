package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/router"
)

// path: /reload
func init() {
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
}
