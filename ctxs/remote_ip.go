package ctxs

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/internal"
	"github.com/zzztttkkk/snow/secret"
)

var forwardSep = []byte(",")

func GetRemoteIp(ctx *fasthttp.RequestCtx) string {
	rip := ctx.Request.Header.Peek("X-Real-IP")
	if len(rip) > 1 {
		return internal.B2s(rip)
	}
	forwards := ctx.Request.Header.Peek("X-Forwarded-For")
	if len(forwards) > 1 {
		return internal.B2s(bytes.Split(forwards, forwardSep)[0])
	}
	return ctx.RemoteIP().String()
}

const ipHashKey = ".iph"

func GetRemoteIpHash(ctx *fasthttp.RequestCtx) string {
	v := ctx.UserValue(ipHashKey)
	if v != nil {
		return v.(string)
	}
	v = internal.B2s(secret.Md5CalcWithSecret(internal.S2b(GetRemoteIp(ctx))))
	ctx.SetUserValue(ipHashKey, v)
	return v.(string)
}
