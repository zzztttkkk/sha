package middleware

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"

	"github.com/zzztttkkk/suna/rbac"

	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/output"
)

var PermissionDeniedError = output.NewError(fasthttp.StatusForbidden, -1, "permission denied")

func NewPermissionCheckMiddleware(policy rbac.CheckPolicy, permissions []string) fasthttp.RequestHandler {
	return NewPermissionCheckHandler(policy, permissions, router.Next)
}

func NewPermissionCheckHandler(
	policy rbac.CheckPolicy,
	permissions []string,
	next fasthttp.RequestHandler,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		user := ctxs.User(ctx)
		if user == nil {
			output.StdError(ctx, fasthttp.StatusUnauthorized)
			return
		}

		if !rbac.IsGranted(ctx, user, policy, permissions...) {
			output.Error(ctx, PermissionDeniedError)
			return
		}

		next(ctx)
	}
}
