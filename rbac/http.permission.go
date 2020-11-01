package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	rpkg "github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

func newPAllPermChecker(perm string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return NewPermissionCheckHandler(PolicyAll, []string{EnsurePermission(perm, "")}, next)
}

// post: /permissions
func init() {
	dig.Append(
		func(loader *rpkg.Loader) {
			loader.Http(
				func(router rpkg.Router) {
					type Form struct {
						Name  string `validate:"L<1-200>"`
						Descp string `validate:"L<1-200>"`
					}

					router.POSTWithDoc(
						"/permissions",
						newPAllPermChecker(
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
		},
	)
}

// delete: /permissions/{pname}
func init() {
	dig.Append(
		func(loader *rpkg.Loader) {
			loader.Http(
				func(router rpkg.Router) {
					type Form struct {
						Name string `validator:"P<name>;L<1-200>"`
					}

					router.DELETEWithDoc(
						"/permissions/{name}",
						newPAllPermChecker(
							"rbac.perm.delete",
							func(ctx *fasthttp.RequestCtx) {
								form := Form{}
								if !validator.Validate(ctx, form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := DelPermission(txc, form.Name); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.GetRules(Form{}).NewDoc(""),
					)
				},
			)
		},
	)
}

// get: /permissions
func init() {
	dig.Append(
		func(loader *rpkg.Loader) {
			loader.Http(
				func(router rpkg.Router) {
					router.GETWithDoc(
						"/permissions",
						newPAllPermChecker(
							"rbac.perm.read",
							func(ctx *fasthttp.RequestCtx) {
								output.MsgOK(ctx, _PermissionOperator.List(ctx))
							},
						),
						rpkg.NewDoc("list all permissions"),
					)
				},
			)
		},
	)
}
