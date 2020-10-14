package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

// path: /role/create
func init() {
	loader.Http(
		func(router router.Router) {
			type Form struct {
				Name  string `validate:"L<1-200>"`
				Descp string `validate:"L<1-200>"`
			}

			router.POSTWithDoc(
				"/role/create",
				newPAllPermChecker(
					"rbac.role.create",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := NewRole(txc, form.Name, form.Descp); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)
		},
	)
}

// path: /role/delete
func init() {
	loader.Http(
		func(router router.Router) {
			type Form struct {
				Name string `validate:"L<1-200>"`
			}

			router.POSTWithDoc(
				"/role/delete",
				newPAllPermChecker(
					"rbac.role.delete",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := DelRole(txc, form.Name); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)
		},
	)
}

// path: /role/list
func init() {
	loader.Http(
		func(R router.Router) {
			R.GETWithDoc(
				"/role/list",
				newPAllPermChecker(
					"rbac.role.list",
					func(ctx *fasthttp.RequestCtx) {
						output.MsgOK(ctx, _RoleOperator.List(ctx))
					},
				),
				router.NewDoc(""),
			)
		},
	)
}

type _NameForm struct {
	Name string `validate:"L<3-255>"`
}

// path: /role/inherit/add
// path: /role/inherit/del
// path: /role/inherit/list
func init() {
	loader.Http(
		func(R router.Router) {
			type Form struct {
				Name  string `validate:"L<1-200>"`
				Based string `validate:"L<1-200>"`
			}

			R.POSTWithDoc(
				"/role/inherit/add",
				newPAllPermChecker(
					"rbac.role.inherit.add",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := RoleAddBased(txc, form.Name, form.Based); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)

			R.POSTWithDoc(
				"/role/inherit/del",
				newPAllPermChecker(
					"rbac.role.inherit.del",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := RoleDelBased(txc, form.Name, form.Based); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)

			R.GETWithDoc(
				"/role/inherit/list",
				newPAllPermChecker(
					"rbac.role.inherit.list",
					func(ctx *fasthttp.RequestCtx) {
						form := _NameForm{}
						if !validator.Validate(ctx, &form) {
							return
						}
						lst, err := RoleListAllBased(ctx, form.Name)
						if err != nil {
							output.Error(ctx, err)
							return
						}
						output.MsgOK(ctx, lst)
					},
				),
				validator.GetRules(_NameForm{}).NewDoc(""),
			)
		},
	)
}

// path: /role/perm/add
// path: /role/perm/del
// path: /role/perm/list
func init() {
	loader.Http(
		func(R router.Router) {
			type Form struct {
				Name string `validate:"L<1-200>"`
				Perm string `validate:"L<1-200>"`
			}

			R.POSTWithDoc(
				"/role/perm/add",
				newPAllPermChecker(
					"rbac.role.add_perm",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := RoleAddPerm(txc, form.Name, form.Perm); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)

			R.POSTWithDoc(
				"/role/perm/del",
				newPAllPermChecker(
					"rbac.role.del_perm",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := RoleDelPerm(txc, form.Name, form.Perm); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)
		},
	)
}
