package middleware

import (
	"fmt"

	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"

	"github.com/zzztttkkk/suna/rbac"

	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/output"
)

type _PermissionPolicyT int

const (
	ALL = _PermissionPolicyT(iota)
	ANY
)

var PermissionDeniedError = output.NewHttpError(fasthttp.StatusForbidden, -1, "permission denied")

func NewPermissionCheckMiddleware(policy _PermissionPolicyT, permissions []string) fasthttp.RequestHandler {
	return NewPermissionCheckHandler(policy, permissions, router.Next)
}

func NewPermissionCheckHandler(
	policy _PermissionPolicyT,
	permissions []string,
	next fasthttp.RequestHandler,
) fasthttp.RequestHandler {
	if policy == ANY {
		return func(ctx *fasthttp.RequestCtx) {
			user := ctxs.Subject(ctx)
			if user == nil {
				output.StdError(ctx, fasthttp.StatusUnauthorized)
				return
			}

			for _, permission := range permissions {
				if rbac.IsGranted(ctx, user, permission) {
					next(ctx)
					return
				}
			}

			output.Error(ctx, PermissionDeniedError)
		}
	}

	if policy != ALL {
		panic(fmt.Errorf("suna.middleware.rbac: unknown policy `%d`", policy))
	}

	return func(ctx *fasthttp.RequestCtx) {
		user := ctxs.Subject(ctx)
		if user == nil {
			output.StdError(ctx, fasthttp.StatusUnauthorized)
			return
		}

		for _, permission := range permissions {
			if !rbac.IsGranted(ctx, user, permission) {
				output.Error(ctx, PermissionDeniedError)
				return
			}
		}

		next(ctx)
	}
}
