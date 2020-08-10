package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
)

type _PermCreateForm struct {
	Name  string `validate:"L<1-200>"`
	Descp string `validate:"L<1-200>"`
}
type _PermDeleteForm struct {
	Name string `validate:"L<1-200>"`
}

func _PermCreateHandler(ctx *fasthttp.RequestCtx) {
	form := _PermCreateForm{}
	if !validator.Validate(ctx, &form) {
		return
	}

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := NewPermission(txc, form.Name, form.Descp); err != nil {
		output.Error(ctx, err)
	}
}

func _PermDeleteHandler(ctx *fasthttp.RequestCtx) {
	form := _PermDeleteForm{}
	if !validator.Validate(ctx, &form) {
		return
	}

	txc, committer := sqls.TxByUser(ctx)
	defer func() { go reload() }()
	defer committer()

	if err := DelPermission(txc, form.Name); err != nil {
		output.Error(ctx, err)
	}
}

func newPermChecker(perm string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	perm = EnsurePermission(perm, "")
	return NewPermissionCheckHandler(PolicyAll, []string{perm}, next)
}

func init() {
	loader.Http(
		func(router router.Router) {
			router.POST(
				"/permission/create",
				newPermChecker(
					"admin.rbac.perm.create",
					_PermCreateHandler,
				),
			)

			router.POST(
				"/permission/delete",
				newPermChecker(
					"admin.rbac.perm.delete",
					_PermDeleteHandler,
				),
			)

			router.GET(
				"/permission/all",
				newPermChecker(
					"admin.rbac.perm.read",
					func(ctx *fasthttp.RequestCtx) {
						output.MsgOK(ctx, _PermissionOperator.List(ctx))
					},
				),
			)
		},
	)
}
