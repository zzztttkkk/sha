package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

func init() {
	loader.Http(
		func(router router.Router) {
			type UserRoleForm struct {
				Uid  int64
				Role string `validate:"L<1-200>"`
			}

			router.POSTWithDoc(
				"/user/role/add",
				newPermChecker(
					"rbac.user.add_role",
					func(ctx *fasthttp.RequestCtx) {
						form := UserRoleForm{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := UserAddRole(txc, form.Uid, form.Role); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(UserRoleForm{}).NewDoc(""),
			)

			router.POSTWithDoc(
				"/user/role/del",
				newPermChecker(
					"rbac.user.del_role",
					func(ctx *fasthttp.RequestCtx) {
						form := UserRoleForm{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := UserDelRole(txc, form.Uid, form.Role); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(UserRoleForm{}).NewDoc(""),
			)

		},
	)
}

func init() {
	loader.Http(
		func(router router.Router) {
			type UidForm struct {
				Uid int64
			}

			router.GETWithDoc(
				"/user/role/list",
				newPermChecker(
					"rbac.user.list_role",
					func(ctx *fasthttp.RequestCtx) {
						form := UidForm{}
						if !validator.Validate(ctx, &form) {
							return
						}
						output.MsgOK(ctx, _UserOperator.getRoles(ctx, form.Uid))
					},
				),
				validator.GetRules(UidForm{}).NewDoc(""),
			)
		},
	)
}
