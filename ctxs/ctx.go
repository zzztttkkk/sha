package ctxs

import (
	"context"
	"errors"
	"github.com/valyala/fasthttp"
)

type _CtxKey int

var _rctxKey = _CtxKey(1111)

func Wrap(ctx *fasthttp.RequestCtx) context.Context {
	return context.WithValue(ctx, _rctxKey, ctx)
}

func Unwrap(ctx context.Context) *fasthttp.RequestCtx {
	v := ctx.Value(_rctxKey)
	if v == nil {
		return nil
	}
	return v.(*fasthttp.RequestCtx)
}

func MustUnwrap(ctx context.Context) *fasthttp.RequestCtx {
	v := Unwrap(ctx)
	if v == nil {
		panic(errors.New("suna.ctxs: empty *fasthttp.RequestCtx"))
	}
	return v
}
