package ctxs

import (
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/internal"
	"github.com/zzztttkkk/snow/middleware/interfaces"
)

func Session(ctx *fasthttp.RequestCtx) interfaces.Session {
	si, ok := ctx.UserValue(internal.RCtxKeySession).(interfaces.Session)
	if !ok {
		return nil
	}
	return si
}
