package middleware

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/secret"
)

type SignOption struct {
	HeaderName string
	Hash       *secret.Hash
}

func NewSignHandler(option *SignOption, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			ctx.Response.Header.SetBytesV(option.HeaderName, option.Hash.Calc(ctx.Response.Body()))
		}()
		next(ctx)
	}
}

func NewSignMiddleware(option *SignOption) fasthttp.RequestHandler {
	return NewSignHandler(option, router.Next)
}
