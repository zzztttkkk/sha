package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
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

	txc, committer := sqls.TxByUser(ctx)
	defer committer()

	if err := UserAddRole(txc, form.Uid, form.Role); err != nil {
		output.Error(ctx, err)
	}
}

func _UserDelRoleHandler(ctx *fasthttp.RequestCtx) {
	form := _UserRoleForm{}
	if !validator.Validate(ctx, &form) {
		return
	}

	txc, committer := sqls.TxByUser(ctx)
	defer committer()

	if err := UserDelRole(txc, form.Uid, form.Role); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/user/role/add",
				newPermChecker("admin.rbac.user.add_role", _UserAddRoleHandler),
			)

			router.POST(
				"/user/role/del",
				newPermChecker("admin.rbac.user.del_role", _UserDelRoleHandler),
			)

			router.GET(
				"/user/role/list",
				newPermChecker(
					"admin.rbac.user.list_role",
					func(ctx *fasthttp.RequestCtx) {
						type UidForm struct {
							Uid int64
						}
						form := UidForm{}
						if !validator.Validate(ctx, &form) {
							return
						}
						output.MsgOK(ctx, _UserOperator.getRoles(ctx, form.Uid))
					},
				),
			)
		},
	)
}
