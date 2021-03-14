package rbac

import "context"

type _CtxKey int

const (
	_RawRequestCtxKey = _CtxKey(iota)
)

func wrapCtx(ctx RCtx) context.Context { return context.WithValue(ctx, _RawRequestCtxKey, ctx) }

func UnwrapRequestCtx(ctx context.Context) context.Context {
	v, ok := ctx.(RCtx)
	if ok {
		return v
	}

	ret := ctx.Value(_RawRequestCtxKey)
	if ret != nil {
		return ret.(context.Context)
	}
	return nil
}
