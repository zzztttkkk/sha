package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
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

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := NewRole(txc, form.Name, form.Descp); err != nil {
		output.Error(ctx, err)
	}
}

func _RoleDeleteHandler(ctx *fasthttp.RequestCtx) {
	form := _RoleDeleteForm{}
	if !validator.Validate(ctx, &form) {
		return
	}

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := DelRole(txc, form.Name); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/role/create",
				newPermChecker("admin.rbac.role.create", _RoleCreateHandler),
			)

			router.POST(
				"/role/delete",
				newPermChecker("admin.rbac.role.delete", _RoleDeleteHandler),
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

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := RoleAddBased(txc, form.Name, form.Based); err != nil {
		output.Error(ctx, err)
	}
}

func _RoleDelBased(ctx *fasthttp.RequestCtx) {
	form := _RoleInheritanceForm{}
	if !validator.Validate(ctx, &form) {
		return
	}

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := RoleDelBased(txc, form.Name, form.Based); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/role/inheritance/add",
				newPermChecker("admin.rbac.role.add_based", _RoleAddBased),
			)

			router.POST(
				"/role/inheritance/del",
				newPermChecker("admin.rbac.role.del_based", _RoleDelBased),
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

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := RoleAddPerm(txc, form.Name, form.Perm); err != nil {
		output.Error(ctx, err)
	}
}

func _RoleDelPerm(ctx *fasthttp.RequestCtx) {
	form := _RolePermForm{}
	if !validator.Validate(ctx, &form) {
		return
	}

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := RoleDelPerm(txc, form.Name, form.Perm); err != nil {
		output.Error(ctx, err)
	}
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/role/perm/add",
				newPermChecker("admin.rbac.role.add_perm", _RoleAddPerm),
			)

			router.POST(
				"/role/perm/del",
				newPermChecker("admin.rbac.role.del_perm", _RoleDelPerm),
			)

			router.GET(
				"/role/all",
				newPermChecker(
					"admin.rbac.role.list",
					func(ctx *fasthttp.RequestCtx) {
						output.MsgOK(ctx, _RoleOperator.List(ctx))
					},
				),
			)
		},
	)
}
