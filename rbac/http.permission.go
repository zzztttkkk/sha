package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	rpkg "github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

func newPermChecker(perm string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	perm = EnsurePermission(perm, "")
	return NewPermissionCheckHandler(PolicyAll, []string{perm}, next)
}

// path: /permission/create
func init() {
	loader.Http(
		func(router rpkg.Router) {
			type Form struct {
				Name  string `validate:"L<1-200>"`
				Descp string `validate:"L<1-200>"`
			}

			router.POSTWithDoc(
				"/permission/create",
				newPermChecker(
					"rbac.perm.create",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := NewPermission(txc, form.Name, form.Descp); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(Form{}).NewDoc(""),
			)
		},
	)
}

// path: /permission/delete
func init() {
	loader.Http(
		func(router rpkg.Router) {
			router.POSTWithDoc(
				"/permission/delete",
				newPermChecker(
					"rbac.perm.delete",
					func(ctx *fasthttp.RequestCtx) {
						form := _NameForm{}
						if !validator.Validate(ctx, &form) {
							return
						}

						txc, committer := sqls.TxByUser(ctx)
						defer committer()

						if err := DelPermission(txc, form.Name); err != nil {
							output.Error(ctx, err)
						}
					},
				),
				validator.GetRules(_NameForm{}).NewDoc(""),
			)
		},
	)
}

// path: /permission/all
func init() {
	loader.Http(
		func(router rpkg.Router) {
			router.GETWithDoc(
				"/permission/all",
				newPermChecker(
					"rbac.perm.read",
					func(ctx *fasthttp.RequestCtx) {
						output.MsgOK(ctx, _PermissionOperator.List(ctx))
					},
				),
				rpkg.NewDoc(""),
			)
		},
	)
}
