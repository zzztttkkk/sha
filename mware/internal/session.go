package internal

import "github.com/valyala/fasthttp"

type Session interface {
	Get(key string, dist interface{}) bool
	Set(key string, val interface{})
	Del(keys ...string)
	CaptchaGenerate(ctx *fasthttp.RequestCtx)
	CaptchaVerify(ctx *fasthttp.RequestCtx) (ok bool)
}
