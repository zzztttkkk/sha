package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

// post: /roles
func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(router router.Router) {
					type Form struct {
						Name  string `validate:"L<1-200>"`
						Descp string `validate:"L<1-200>"`
					}

					router.POSTWithDoc(
						"/roles",
						newPAllPermChecker(
							"rbac.roles.create",
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
		},
	)
}

// delete: /roles/<rname>
func init() {
	dig.Append(
		func(loader *router.Loader) {
			type Form struct {
				Name string `validator:"P<name>;L<1-200>"`
			}

			loader.Http(
				func(router router.Router) {
					router.DELETE(
						"/roles/{name}",
						newPAllPermChecker(
							"rbac.roles.delete",
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
					)
				},
			)
		},
	)
}

// get: /roles
func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(R router.Router) {
					R.GET(
						"/roles",
						newPAllPermChecker(
							"rbac.role.list",
							func(ctx *fasthttp.RequestCtx) {
								output.MsgOK(ctx, _RoleOperator.List(ctx))
							},
						),
					)
				},
			)

		},
	)
}

// post: /roles/{name}/inherits
// delete: /roles/{rname}/inherits/{iname}
// get: /roles/{name}/inherits
func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(R router.Router) {
					type CreateForm struct {
						RoleName string `validator:"P<name>;L<1-200>"`
						Name     string `validator:"L<1-200>"`
					}

					R.POSTWithDoc(
						"/roles/{name}/inherits",
						newPAllPermChecker(
							"rbac.role.inherits.add",
							func(ctx *fasthttp.RequestCtx) {
								form := CreateForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := RoleInheritFrom(txc, form.RoleName, form.Name); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.MakeDoc(CreateForm{}, ""),
					)

					type DeleteForm struct {
						RoleName  string `validator:"P<rname>;L<1-200>"`
						BasedName string `validator:"P<iname>;L<1-200>"`
					}

					R.DELETEWithDoc(
						"/roles/{rname}/inherits/{iname}",
						newPAllPermChecker(
							"rbac.role.inherits.del",
							func(ctx *fasthttp.RequestCtx) {
								form := DeleteForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := RoleUninheritFrom(txc, form.RoleName, form.BasedName); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.MakeDoc(DeleteForm{}, ""),
					)

					type ListForm struct {
						RoleName string `validator:"P<rname>;L<1-200>"`
					}

					R.GETWithDoc(
						"/roles/{rname}/inherits",
						newPAllPermChecker(
							"rbac.role.inherits.list",
							func(ctx *fasthttp.RequestCtx) {
								form := ListForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								lst, err := RoleListAllBased(ctx, form.RoleName)
								if err != nil {
									output.Error(ctx, err)
									return
								}
								output.MsgOK(ctx, lst)
							},
						),
						validator.MakeDoc(ListForm{}, ""),
					)
				},
			)
		},
	)
}

// post: /roles/{name}/perms
// delete: /roles/{rname}/perms/{pname}
// get: /roles/{name}/perms
func init() {
	dig.Append(
		func(loader *router.Loader) {
			loader.Http(
				func(R router.Router) {
					type CreateForm struct {
						RoleName string `validator:"P<name>;L<1-200>"`
						Name     string `validator:"L<1-200>"`
					}

					R.POSTWithDoc(
						"/roles/{name}/perms",
						newPAllPermChecker(
							"rbac.role.perms.add",
							func(ctx *fasthttp.RequestCtx) {
								form := CreateForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := RoleAddPerm(txc, form.RoleName, form.Name); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.MakeDoc(CreateForm{}, ""),
					)

					type DeleteForm struct {
						RoleName string `validator:"P<rname>;L<1-200>"`
						PermName string `validator:"P<pname>;L<1-200>"`
					}

					R.DELETEWithDoc(
						"/roles/{rname}/perms/{pname}",
						newPAllPermChecker(
							"rbac.role.perms.del",
							func(ctx *fasthttp.RequestCtx) {
								form := DeleteForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								txc, committer := sqls.TxByUser(ctx)
								defer committer()

								if err := RoleDelPerm(txc, form.RoleName, form.PermName); err != nil {
									output.Error(ctx, err)
								}
							},
						),
						validator.MakeDoc(DeleteForm{}, ""),
					)

					type ListForm struct {
						RoleName string `validator:"P<rname>;L<1-200>"`
					}

					R.GETWithDoc(
						"/roles/{rname}/perms",
						newPAllPermChecker(
							"rbac.role.perms.list",
							func(ctx *fasthttp.RequestCtx) {
								form := ListForm{}
								if !validator.Validate(ctx, &form) {
									return
								}

								role, ok := _RoleOperator.GetByName(ctx, form.RoleName)
								if !ok {
									output.Error(ctx, output.HttpErrors[fasthttp.StatusNotFound])
									return
								}

								var perms []sqls.EnumItem
								for _, pid := range role.(*Role).Permissions {
									p, ok := _PermissionOperator.GetById(ctx, pid)
									if !ok {
										continue
									}
									perms = append(perms, p)
								}
								output.MsgOK(ctx, perms)
							},
						),
						validator.MakeDoc(ListForm{}, ""),
					)
				},
			)
		},
	)
}
