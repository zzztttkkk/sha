package sha

import (
	"github.com/zzztttkkk/sha/rbac"
	"github.com/zzztttkkk/sha/validator"
)

type _RbacRouterAdapter struct {
	Router
}

func (r *_RbacRouterAdapter) HandleWithDoc(
	method string, path string,
	handler rbac.HandlerFunc,
	doc validator.Document,
) {
	r.Router.HTTPWithDocument(
		method,
		path,
		RequestHandlerFunc(func(ctx *RequestCtx) { handler(ctx) }),
		doc,
	)
}

func UseRBAC(router Router, options *rbac.Options) { rbac.Init(&_RbacRouterAdapter{router}, options) }

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
