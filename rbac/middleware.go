package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	"log"
)

var ErrPermissionDenied = output.NewError(fasthttp.StatusForbidden, -1, "permission denied")

func NewPermissionCheckMiddleware(policy CheckPolicy, permissions ...string) fasthttp.RequestHandler {
	return NewPermissionCheckHandler(policy, permissions, router.Next)
}

func NewPermissionCheckHandler(
	policy CheckPolicy,
	permissions []string,
	next fasthttp.RequestHandler,
) fasthttp.RequestHandler {
	for _, p := range permissions {
		if !_PermissionOperator.ExistsByName(context.Background(), p) {
			log.Fatalf("suna.rbac: permission `%s` is not exists", p)
		}
	}

	return func(ctx *fasthttp.RequestCtx) {
		user := auth.MustGetUser(ctx)
		if !IsGranted(ctx, user, policy, permissions...) {
			output.Error(ctx, ErrPermissionDenied)
			return
		}
		next(ctx)
	}
}
