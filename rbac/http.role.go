package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/validator"
)

type _RoleCreateForm struct {
	Name  string `validate:"L<1-200>"`
	Descp string `validate:"L<1-200>"`
}
type _RoleDeleteForm struct {
	Name string `validate:"L<1-200>"`
}

func _RoleCreateHandler(ctx *fasthttp.RequestCtx) {
	form := _RoleCreateForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := NewRole(ctxs.Std(ctx), form.Name, form.Descp); err != nil {
		output.Error(ctx, err)
	}
}

func _RoleDeleteHandler(ctx *fasthttp.RequestCtx) {
	form := _RoleDeleteForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := DelRole(ctxs.Std(ctx), form.Name); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/role/create",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.role.create", "")},
					_RoleCreateHandler,
				),
			)

			router.POST(
				"/role/delete",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.role.delete", "")},
					_RoleDeleteHandler,
				),
			)
		},
	)
}

type _RoleInheritanceForm struct {
	Name  string `validate:"L<1-200>"`
	Based string `validate:"L<1-200>"`
}

func _RoleAddBased(ctx *fasthttp.RequestCtx) {
	form := _RoleInheritanceForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := RoleAddBased(ctxs.Std(ctx), form.Name, form.Based); err != nil {
		output.Error(ctx, err)
	}
}

func _RoleDelBased(ctx *fasthttp.RequestCtx) {
	form := _RoleInheritanceForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := RoleDelBased(ctxs.Std(ctx), form.Name, form.Based); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/role/inheritance/add",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.role.add_based", "")},
					_RoleAddBased,
				),
			)

			router.POST(
				"/role/inheritance/del",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.role.del_based", "")},
					_RoleDelBased,
				),
			)
		},
	)
}

type _RolePermForm struct {
	Name string `validate:"L<1-200>"`
	Perm string `validate:"L<1-200>"`
}

func _RoleAddPerm(ctx *fasthttp.RequestCtx) {
	form := _RolePermForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := RoleAddPerm(ctxs.Std(ctx), form.Name, form.Perm); err != nil {
		output.Error(ctx, err)
	}
}

func _RoleDelPerm(ctx *fasthttp.RequestCtx) {
	form := _RolePermForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := RoleDelPerm(ctxs.Std(ctx), form.Name, form.Perm); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/role/perm/add",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.role.add_perm", "")},
					_RoleAddPerm,
				),
			)

			router.POST(
				"/role/perm/del",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.role.del_perm", "")},
					_RoleDelPerm,
				),
			)
		},
	)
}
