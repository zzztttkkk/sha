package ctxs

import (
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/internal"
	"github.com/zzztttkkk/snow/middleware/interfaces"
)

func Uid(ctx *fasthttp.RequestCtx) int64 {
	iv := ctx.UserValue(internal.RCtxKeyUid)
	if iv != nil {
		return iv.(int64)
	}
	return -1
}

func AuthExt(ctx *fasthttp.RequestCtx) map[string]interface{} {
	iv := ctx.UserValue(internal.RCtxKeyTokenExt)
	if iv != nil {
		return iv.(map[string]interface{})
	}
	return nil
}

func LastLogin(ctx *fasthttp.RequestCtx) int64 {
	iv := ctx.UserValue(internal.RCtxKeyLastLogin)
	if iv != nil {
		return iv.(int64)
	}
	return -1
}

func User(ctx *fasthttp.RequestCtx) interfaces.User {
	iv, ok := ctx.UserValue(internal.RCtxKeyUser).(interfaces.User)
	if ok {
		return iv
	}
	return nil
}
