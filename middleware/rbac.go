package middleware

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/middleware/ctxs"
	"github.com/zzztttkkk/snow/output"
	rbacpkg "github.com/zzztttkkk/snow/rbac"
	"github.com/zzztttkkk/snow/router"
)

type _PermissionPolicyT int

const (
	ALL = _PermissionPolicyT(iota)
	ANY
)

var PermissionDeniedError = output.NewHttpError(fasthttp.StatusForbidden, -1, "permission denied")

func NewPermissionCheckMiddleware(rbac rbacpkg.Rbac, policy _PermissionPolicyT, permissions []string) fasthttp.RequestHandler {
	return NewPermissionCheckHandler(rbac, policy, permissions, router.Next)
}

func NewPermissionCheckHandler(
	rbac rbacpkg.Rbac,
	policy _PermissionPolicyT,
	permissions []string,
	next fasthttp.RequestHandler,
) fasthttp.RequestHandler {
	if policy == ANY {
		return func(ctx *fasthttp.RequestCtx) {
			user := ctxs.User(ctx)
			if user == nil {
				output.StdError(ctx, fasthttp.StatusUnauthorized)
				return
			}

			for _, permission := range permissions {
				ok, _ := rbac.IsGranted(ctx, user, permission)
				if ok {
					next(ctx)
					return
				}
			}

			output.Error(ctx, PermissionDeniedError)
		}
}

	if policy != ALL {
		panic(fmt.Errorf("snow.middleware.rbac: unknown policy `%d`", policy))
	}

	return func(ctx *fasthttp.RequestCtx) {
		user := ctxs.User(ctx)
		if user == nil {
			output.StdError(ctx, fasthttp.StatusUnauthorized)
			return
		}

		for _, permission := range permissions {
			ok, _ := rbac.IsGranted(ctx, user, permission)
			if !ok {
				output.Error(ctx, PermissionDeniedError)
				return
			}
		}

		next(ctx)
	}
}
