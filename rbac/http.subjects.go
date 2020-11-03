package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(router router.Router) {
					type AddForm struct {
						Sid      int64  `validator:"P<sid>"`
						RoleName string `validator:"L<1-200>"`
					}

					router.POSTWithDoc(
						"/subjects/{sid}/roles",
						newPAllPermChecker(
							"rbac.subjects.roles.add",
							func(ctx *fasthttp.RequestCtx) {
								form := AddForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := SubjectAddRole(txc, form.Sid, form.RoleName); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.MakeDoc(AddForm{}, ""),
					)

					type DelForm struct {
						Sid      int64  `validator:"P<sid>"`
						RoleName string `validator:"P<rname>;L<1-200>"`
					}

					router.POSTWithDoc(
						"/subjects/{sid}/roles/{rname}",
						newPAllPermChecker(
							"rbac.subjects.roles.del",
							func(ctx *fasthttp.RequestCtx) {
								form := DelForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := SubjectDelRole(txc, form.Sid, form.RoleName); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.MakeDoc(DelForm{}, ""),
					)

				},
			)
		},
	)
}

func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(router router.Router) {
					type UidForm struct {
						Uid int64
					}

					router.GETWithDoc(
						"/subject/{sid}/role/list",
						newPAllPermChecker(
							"rbac.subject.list_role",
							func(ctx *fasthttp.RequestCtx) {
								form := UidForm{}
								if !validator.Validate(ctx, &form) {
									return
								}
								output.MsgOK(ctx, SubjectOperator.getRoles(ctx, form.Uid))
							},
						),
						validator.GetRules(UidForm{}).NewDoc(""),
					)
				},
			)
		},
	)
}
