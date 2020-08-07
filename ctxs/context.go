package ctxs

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
)

type StdCtxKey int

var (
	_RCtxKeyStdCtx = StdCtxKey(0)
)

func Wrap(ctx *fasthttp.RequestCtx) context.Context {
	return context.WithValue(ctx, _RCtxKeyStdCtx, ctx)
}

func Unwrap(ctx context.Context) *fasthttp.RequestCtx {
	v, ok := ctx.Value(_RCtxKeyStdCtx).(*fasthttp.RequestCtx)
	if !ok {
		panic(fmt.Errorf("suna.ctxs: nil fasthttp.Unwrap"))
	}
	return v
}
