package ctxs

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/middleware/interfaces"
	"github.com/zzztttkkk/snow/middleware/internal"
)

func Session(ctx *fasthttp.RequestCtx) interfaces.Session {
	si, ok := ctx.UserValue(internal.SessionKey).(interfaces.Session)
	if !ok {
		return nil
	}
	return si
}
