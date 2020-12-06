package suna

import (
	"github.com/zzztttkkk/suna/rbac"
	"log"
)

type _RbacR struct {
	Router
}

func (r *_RbacR) HandleWithDoc(
	method string, path string,
	handler rbac.HandlerFunc,
	doc interface{},
) {
	r.Router.RESTWithForm(
		method,
		path,
		RequestHandlerFunc(func(ctx *RequestCtx) { handler(ctx, ctx) }),
		doc,
	)
}

func UseRBAC(
	router Router,
	tableNamePrefix string,
	logger *log.Logger,
	loggingFroReadOperation bool,
) {
	rbac.Init(
		&rbac.Options{
			Router:           &_RbacR{router},
			TableNamePrefix:  tableNamePrefix,
			LogReadOperation: loggingFroReadOperation,
			Logger:           logger,
		},
	)
}

func MustGrantedAll(permissions ...string) Middleware {
	return MiddlewareFunc(
		func(ctx *RequestCtx, next func()) {
			err := rbac.GrantedAll(ctx, permissions...)
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
			err := rbac.GrantedAny(ctx, permissions...)
			if err != nil {
				panic(err)
			}
			next()
		},
	)
}
