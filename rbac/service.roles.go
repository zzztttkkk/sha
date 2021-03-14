package rbac

import (
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
)

func init() {
	type Form struct {
		Name string `validate:"name,L=1-512"`
		Desc string `validate:"desc,optional"`
	}

	register(
		"POST",
		"/roles",
		func(rctx RCtx) {
			ctx, committer := sqlx.Tx(wrapCtx(rctx))
			defer committer()

			MustGrantedAll(ctx, PermRoleCreate)

			var form Form
			rctx.MustValidate(&form)
			dao.NewRole(ctx, form.Name, form.Desc)
		},
		Form{},
	)
}

func init() {
	register(
		"GET",
		"/roles",
		func(rctx RCtx) {
			MustGrantedAll(rctx, PermRoleListAll)

			lst := dao.Roles(rctx)
			for _, r := range lst {
				r.Permissions = dao.RolePerms(rctx, r.ID)
			}

			rctx.WriteJSON(lst)
		},
		nil,
	)
}

func init() {
	type Form struct {
		RoleName string `validate:",P=rname,L=1-512"`
	}

	register(
		"GET",
		"/role/:rname",
		func(rctx RCtx) {
			var form Form
			rctx.MustValidate(&form)

			MustGrantedAny(rctx, "rbac.roles.listAll", "rbac.role."+form.RoleName+".read")

			role := dao.RoleByName(rctx, form.RoleName)

			if role == nil {
				rctx.SetStatus(http.StatusNotFound)
				return
			}

			role.Permissions = dao.RolePerms(rctx, role.ID)
			rctx.WriteJSON(role)
		},
		nil,
	)
}

// post /role/:rname/perms 	grant one perm to role
func init() {
	type Form struct {
		RoleName string `validate:",P=rname"`
		Name     string `validate:"name,L=1-512"`
	}

	internal.Dig.Append(
		func(router Router, _ internal.DaoOK) {
			router.HandleWithDoc(
				"POST",
				"/role/:rname/perms",
				func(rctx RCtx) {
					ctx, committer := sqlx.Tx(wrapCtx(rctx))
					defer committer()

					MustGrantedAll(ctx, "rbac.roles.create")

					var form Form
					rctx.MustValidate(&form)
					dao.RoleAddPerm(ctx, form.RoleName, form.Name)
				},
				validator.NewMarkdownDocument(Form{}, validator.Undefined),
			)
		},
	)
}

// delete /role/:rname/perms/:pname  cancel pm perm from role
func init() {
	type Form struct {
		RoleName string `validate:",P=rname,L=1-512"`
		PermName string `validate:",P=pname,L=1-512"`
	}

	internal.Dig.Append(
		func(router Router, _ internal.DaoOK) {
			router.HandleWithDoc(
				"DELETE",
				"/role/:rname/perms/:pname",
				func(rctx RCtx) {
					ctx, committer := sqlx.Tx(wrapCtx(rctx))
					defer committer()

					MustGrantedAll(ctx, "rbac.roles.delete")

					var form Form
					rctx.MustValidate(&form)

					dao.RoleDelPerm(ctx, form.RoleName, form.PermName)
				},
				validator.NewMarkdownDocument(Form{}, validator.Undefined),
			)
		},
	)
}
