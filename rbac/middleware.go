package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/output"
)

var PermissionDeniedError = output.NewError(fasthttp.StatusForbidden, -1, "permission denied")

func NewPermissionCheckMiddleware(policy CheckPolicy, permissions []string) fasthttp.RequestHandler {
	return NewPermissionCheckHandler(policy, permissions, router.Next)
}

func NewPermissionCheckHandler(
	policy CheckPolicy,
	permissions []string,
	next fasthttp.RequestHandler,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		user := auth.GetUserMust(ctx)
		if !IsGranted(ctx, user, policy, permissions...) {
			output.Error(ctx, PermissionDeniedError)
			return
		}

		next(ctx)
	}
}
