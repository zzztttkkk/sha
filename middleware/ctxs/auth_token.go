package ctxs

import (
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/middleware/interfaces"
)

func User(ctx *fasthttp.RequestCtx) interfaces.User {
	iv, ok := ctx.UserValue(internal.RCtxKeyUser).(interfaces.User)
	if ok {
		return iv
	}
	return nil
}
