package ctxs

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/mware/internal"
)

func Session(ctx *fasthttp.RequestCtx) internal.Session {
	si, ok := ctx.UserValue(internal.SessionKey).(internal.Session)
	if !ok {
		return nil
	}
	return si
}
