package rbac

import (
	"context"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
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
		func(rctx context.Context, rw ReqWriter) {
			ctx, committer := sqlx.Tx(rctx)
			defer committer()

			MustGrantedAll(ctx, PermRoleCreate)

			var form Form
			rw.MustValidate(&form)
			dao.NewRole(ctx, form.Name, form.Desc)
		},
		Form{},
	)
}

func init() {
	register(
		"GET",
		"/roles",
		func(rctx context.Context, rw ReqWriter) {
			MustGrantedAll(rctx, PermRoleListAll)

			lst := dao.Roles(rctx)
			for _, r := range lst {
				r.Permissions = dao.RolePerms(rctx, r.ID)
			}

			rw.WriteJSON(lst)
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
		func(ctx context.Context, rw ReqWriter) {
			var form Form
			rw.MustValidate(&form)

			MustGrantedAny(ctx, "rbac.roles.listAll", "rbac.role."+form.RoleName+".read")

			role := dao.RoleByName(ctx, form.RoleName)

			if role == nil {
				rw.SetStatus(http.StatusNotFound)
				return
			}

			role.Permissions = dao.RolePerms(ctx, role.ID)
			rw.WriteJSON(role)
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
				func(rctx context.Context, rw ReqWriter) {
					ctx, committer := sqlx.Tx(rctx)
					defer committer()

					MustGrantedAll(ctx, "rbac.roles.create")

					var form Form
					rw.MustValidate(&form)
					dao.RoleAddPerm(ctx, form.RoleName, form.Name)
				},
				Form{},
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
				func(rctx context.Context, rw ReqWriter) {
					ctx, committer := sqlx.Tx(rctx)
					defer committer()

					MustGrantedAll(ctx, "rbac.roles.delete")

					var form Form
					rw.MustValidate(&form)

					dao.RoleDelPerm(ctx, form.RoleName, form.PermName)
				},
				Form{},
			)
		},
	)
}
