package suna

import (
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/rbac/auth"
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
	authenticator auth.Authenticator,
	tableNamePrefix string,
	logger *log.Logger,
	loggingFroReadOperation bool,
) {
	rbac.Init(
		&rbac.Options{
			Authenticator:    authenticator,
			Router:           &_RbacR{router},
			TableNamePrefix:  tableNamePrefix,
			LogReadOperation: loggingFroReadOperation,
			Logger:           logger,
		},
	)
}
