package ctxs

import (
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/middleware/interfaces"
)

func Session(ctx *fasthttp.RequestCtx) interfaces.Session {
	si, ok := ctx.UserValue(internal.RCtxKeySession).(interfaces.Session)
	if !ok {
		return nil
	}
	return si
}
