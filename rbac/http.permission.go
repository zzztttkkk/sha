package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/validator"
)

type _PermCreateForm struct {
	Name  string `validate:"L<1-200>"`
	Descp string `validate:"L<1-200>"`
}
type _PermDeleteForm struct {
	Name string `validate:"L<1-200>"`
}

func _PermCreateHandler(ctx *fasthttp.RequestCtx) {
	form := _PermCreateForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := NewPermission(wrapRCtx(ctx), form.Name, form.Descp); err != nil {
		output.Error(ctx, err)
	}
}

func _PermDeleteHandler(ctx *fasthttp.RequestCtx) {
	form := _PermDeleteForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := DelPermission(wrapRCtx(ctx), form.Name); err != nil {
		output.Error(ctx, err)
	}
}

func PermPageHandler(ctx *fasthttp.RequestCtx) {

}

func newPermChecker(perm string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	perm = EnsurePermission(perm, "")

	return func(ctx *fasthttp.RequestCtx) {
		user := getUserFromRCtx(ctx)
		if user == nil {
			output.Error(ctx, output.HttpErrors[fasthttp.StatusUnauthorized])
			return
		}

		if !IsGranted(ctx, user, PolicyAll, perm) {
			output.Error(ctx, output.HttpErrors[fasthttp.StatusForbidden])
			return
		}

		next(ctx)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/permission/create",
				newPermChecker(
					"admin.rbac.perm.create",
					_PermCreateHandler,
				),
			)

			router.POST(
				"/permission/delete",
				newPermChecker(
					"admin.rbac.perm.delete",
					_PermDeleteHandler,
				),
			)

			router.GET(
				"/permission/all",
				newPermChecker(
					"admin.rbac.perm.read",
					func(ctx *fasthttp.RequestCtx) {
						output.MsgOK(ctx, _PermissionOperator.List(ctx))
					},
				),
			)
		},
	)
}
