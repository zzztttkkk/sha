package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/middleware"
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
	if err := NewPermission(ctxs.Std(ctx), form.Name, form.Descp); err != nil {
		output.Error(ctx, err)
	}
}

func _PermDeleteHandler(ctx *fasthttp.RequestCtx) {
	form := _PermDeleteForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := DelPermission(ctxs.Std(ctx), form.Name); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/permission/create",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.permission.create", "")},
					_PermCreateHandler,
				),
			)

			router.POST(
				"/permission/delete",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.permission.delete", "")},
					_PermDeleteHandler,
				),
			)
		},
	)
}
