package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/validator"
)

type _UserRoleForm struct {
	Uid  int64
	Role string `validate:"L<1-200>"`
}

func _UserAddRoleHandler(ctx *fasthttp.RequestCtx) {
	form := _UserRoleForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := UserAddRole(ctxs.Std(ctx), form.Uid, form.Role); err != nil {
		output.Error(ctx, err)
	}
}

func _UserDelRoleHandler(ctx *fasthttp.RequestCtx) {
	form := _UserRoleForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	if err := UserDelRole(ctxs.Std(ctx), form.Uid, form.Role); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/user/role/add",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.user.add_role", "")},
					_UserAddRoleHandler,
				),
			)

			router.POST(
				"/user/role/del",
				middleware.NewPermissionCheckHandler(
					PolicyAll,
					[]string{EnsurePermission("admin.rbac.user.del_role", "")},
					_UserDelRoleHandler,
				),
			)
		},
	)
}
