package ctxs

import (
	"bytes"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/utils"
)

var forwardSep = []byte(",")

func RemoteIp(ctx *fasthttp.RequestCtx) string {
	rip := ctx.Request.Header.Peek("X-Real-IP")
	if len(rip) > 1 {
		return utils.B2s(rip)
	}
	forwards := ctx.Request.Header.Peek("X-Forwarded-For")
	if len(forwards) > 1 {
		return utils.B2s(bytes.Split(forwards, forwardSep)[0])
	}
	return ctx.RemoteIP().String()
}

func RemoteIpHash(ctx *fasthttp.RequestCtx) string {
	v := ctx.UserValue(internal.RCtxKeyRemoteIp)
	if v != nil {
		return v.(string)
	}
	v = utils.B2s(secret.Md5.Calc(utils.S2b(RemoteIp(ctx))))
	ctx.SetUserValue(internal.RCtxKeyRemoteIp, v)
	return v.(string)
}
