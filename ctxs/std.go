package ctxs

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
)

func Std(ctx *fasthttp.RequestCtx) context.Context {
	return context.WithValue(ctx, internal.RCtxKeyStdCtx, ctx)
}

func RequestCtx(ctx context.Context) *fasthttp.RequestCtx {
	v, ok := ctx.Value(internal.RCtxKeyStdCtx).(*fasthttp.RequestCtx)
	if !ok {
		panic(fmt.Errorf("suna.ctxs: nil fasthttp.RequestCtx"))
	}
	return v
}
