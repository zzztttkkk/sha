package ctxs

import (
	"context"
	"errors"
	"github.com/valyala/fasthttp"
)

type _CtxKey int

var _RCtxKey = _CtxKey(1111)

func Wrap(ctx *fasthttp.RequestCtx) context.Context {
	return context.WithValue(ctx, _RCtxKey, ctx)
}

func Unwrap(ctx context.Context) *fasthttp.RequestCtx {
	v := ctx.Value(_RCtxKey)
	if v == nil {
		panic(errors.New("suna.ctxs: empty *fasthttp.RequestCtx"))
	}
	return v.(*fasthttp.RequestCtx)
}
