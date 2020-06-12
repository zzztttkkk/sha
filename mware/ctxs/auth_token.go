package ctxs

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/mware/internal"
)

func Uid(ctx *fasthttp.RequestCtx) int64 {
	iv := ctx.UserValue(internal.Uid)
	if iv != nil {
		return iv.(int64)
	}
	return -1
}

func AuthExt(ctx *fasthttp.RequestCtx) map[string]interface{} {
	iv := ctx.UserValue(internal.TokenExt)
	if iv != nil {
		return iv.(map[string]interface{})
	}
	return nil
}

func LastLogin(ctx *fasthttp.RequestCtx) int64 {
	iv := ctx.UserValue(internal.LastLogin)
	if iv != nil {
		return iv.(int64)
	}
	return -1
}

func User(ctx *fasthttp.RequestCtx) mware.User {
	iv, ok := ctx.UserValue(internal.UserKey).(mware.User)
	if ok {
		return iv
	}
	return nil
}
