package sha

import (
	"context"
	"github.com/zzztttkkk/sha/rbac"
	"github.com/zzztttkkk/sha/validator"
)

type _RbacRouterAdapter struct {
	Router
}

var _ rbac.Router = (*_RbacRouterAdapter)(nil)

func (r *_RbacRouterAdapter) HTTP(method string, path string, handler rbac.HandlerFunc, doc validator.Document) {
	r.Router.HTTPWithOptions(
		&HandlerOptions{Document: doc},
		method, path,
		RequestHandlerFunc(func(ctx *RequestCtx) { handler(Wrap(ctx)) }),
	)
}

type _CtxAdapter struct{}

func (_CtxAdapter) ValidateForm(ctx context.Context, dist interface{}) error {
	return Unwrap(ctx).ValidateForm(dist)
}

func (_CtxAdapter) SetResponseStatus(ctx context.Context, v int) {
	Unwrap(ctx).Response.statusCode = v
}

func (_CtxAdapter) Write(ctx context.Context, p []byte) (int, error) {
	return Unwrap(ctx).Response.Write(p)
}

func (_CtxAdapter) WriteJSON(ctx context.Context, v interface{}) {
	Unwrap(ctx).WriteJSON(v)
}

func (_CtxAdapter) SetError(ctx context.Context, v interface{}) {
	Unwrap(ctx).err = v
}

var _ rbac.CtxAdapter = (*_CtxAdapter)(nil)

func UseRBAC(router Router, options *rbac.Options) {
	rbac.Init(&_RbacRouterAdapter{router}, _CtxAdapter{}, options)
}

func MustGrantedAll(permissions ...string) Middleware {
	return MiddlewareFunc(
		func(ctx *RequestCtx, next func()) {
			err := rbac.IsGrantedAll(ctx, permissions...)
			if err != nil {
				panic(err)
			}
			next()
		},
	)
}

func MustGrantedAny(permissions ...string) Middleware {
	return MiddlewareFunc(
		func(ctx *RequestCtx, next func()) {
			err := rbac.IsGrantedAny(ctx, permissions...)
			if err != nil {
				panic(err)
			}
			next()
		},
	)
}
