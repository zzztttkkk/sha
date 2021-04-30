package rbac

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
)

func init() {
	type Form struct {
		Name string `validator:"name,L=1-512,r=rbacname"`
		Desc string `validator:"desc,optional"`
	}

	register(
		"POST",
		"/roles",
		func(ctx context.Context) {
			ctx, tx := sqlx.Tx(ctx)
			defer tx.AutoCommit(ctx)

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
		RoleName string `validator:"rname,w=url,l=1-512,r=rbacname"`
	}

	register(
		"GET",
		"/role/{rname}",
		func(ctx context.Context) {
			var form Form
			if err := gAdapter.ValidateForm(ctx, &form); err != nil {
				gAdapter.SetError(ctx, err)
				return
			}

			MustGrantedAny(ctx, PermRoleListAll, fmt.Sprintf("%s.list", form.RoleName))

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
		RoleName string `validator:"rname,w=url,r=rbacname"`
		Name     string `validator:"name,l=1-512,r=rbacname"`
	}

	internal.Dig.Append(
		func(router Router, _ internal.DaoOK) {
			router.HTTP(
				"POST",
				"/role/{rname}/perms",
				func(ctx context.Context) {
					ctx, tx := sqlx.Tx(ctx)
					defer tx.AutoCommit(ctx)

					MustGrantedAll(ctx, PermRoleAddPerm)

					var form Form
					if err := gAdapter.ValidateForm(ctx, &form); err != nil {
						gAdapter.SetError(ctx, err)
						return
					}

					dao.RoleAddPerm(ctx, form.RoleName, form.Name)
				},
				validator.NewDocument(Form{}, validator.Undefined),
			)
		},
	)
}

// delete /role/:rname/perms/:pname  cancel pm perm from role
func init() {
	type Form struct {
		RoleName string `validate:"rname.w=url,L=1-512,,r=rbacname"`
		PermName string `validate:"pname,w=url,L=1-512,,r=rbacname"`
	}

	internal.Dig.Append(
		func(router Router, _ internal.DaoOK) {
			router.HTTP(
				"DELETE",
				"/role/{rname}/perms/{pname}",
				func(ctx context.Context) {
					ctx, tx := sqlx.Tx(ctx)
					defer tx.AutoCommit(ctx)

					MustGrantedAll(ctx, PermRoleDelPerm)

					var form Form
					if err := gAdapter.ValidateForm(ctx, &form); err != nil {
						gAdapter.SetError(ctx, err)
						return
					}
					dao.RoleDelPerm(ctx, form.RoleName, form.PermName)
				},
				validator.NewDocument(Form{}, validator.Undefined),
			)
		},
	)
}
