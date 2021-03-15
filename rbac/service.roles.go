package rbac

import (
	"context"
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
		func(ctx context.Context) {
			ctx, committer := sqlx.Tx(ctx)
			defer committer()

			MustGrantedAll(ctx, PermRoleCreate)

			var form Form
			if err := gAdapter.ValidateForm(ctx, &form); err != nil {
				gAdapter.SetError(ctx, err)
				return
			}
			dao.NewRole(ctx, form.Name, form.Desc)
		},
		Form{},
	)
}

func init() {
	register(
		"GET",
		"/roles",
		func(ctx context.Context) {
			MustGrantedAll(ctx, PermRoleListAll)
			lst := dao.Roles(ctx)
			for _, r := range lst {
				r.Permissions = dao.RolePerms(ctx, r.ID)
			}
			gAdapter.WriteJSON(ctx, lst)
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
		func(ctx context.Context) {
			var form Form
			if err := gAdapter.ValidateForm(ctx, &form); err != nil {
				gAdapter.SetError(ctx, err)
				return
			}

			MustGrantedAny(ctx, "rbac.roles.listAll", "rbac.role."+form.RoleName+".read")

			role := dao.RoleByName(ctx, form.RoleName)

			if role == nil {
				gAdapter.SetResponseStatus(ctx, http.StatusNotFound)
				return
			}

			role.Permissions = dao.RolePerms(ctx, role.ID)
			gAdapter.WriteJSON(ctx, role)
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
			router.HTTP(
				"POST",
				"/role/:rname/perms",
				func(ctx context.Context) {
					ctx, committer := sqlx.Tx(ctx)
					defer committer()

					MustGrantedAll(ctx, "rbac.roles.create")

					var form Form
					if err := gAdapter.ValidateForm(ctx, &form); err != nil {
						gAdapter.SetError(ctx, err)
						return
					}

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
			router.HTTP(
				"DELETE",
				"/role/:rname/perms/:pname",
				func(ctx context.Context) {
					ctx, committer := sqlx.Tx(ctx)
					defer committer()

					MustGrantedAll(ctx, "rbac.roles.delete")

					var form Form
					if err := gAdapter.ValidateForm(ctx, &form); err != nil {
						gAdapter.SetError(ctx, err)
						return
					}
					dao.RoleDelPerm(ctx, form.RoleName, form.PermName)
				},
				validator.NewMarkdownDocument(Form{}, validator.Undefined),
			)
		},
	)
}
