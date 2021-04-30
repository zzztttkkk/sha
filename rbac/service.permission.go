package rbac

import (
	"context"
	"database/sql"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
)

func register(method, path string, fn HandlerFunc, form interface{}) {
	internal.Dig.Append(
		func(router Router, _ _PermOK) {
			router.HTTP(
				method, path,
				func(ctx context.Context) {
					defer Recover(ctx)
					fn(ctx)
				},
				validator.NewDocument(form, validator.Undefined),
			)
		},
	)
}

func Recover(ctx context.Context) {
	v := recover()
	if v == nil {
		return
	}

	switch tv := v.(type) {
	case error:
		switch tv {
		case auth.ErrUnauthenticatedOperation:
			gAdapter.SetResponseStatus(ctx, http.StatusUnauthorized)
			return
		case ErrPermissionDenied:
			gAdapter.SetResponseStatus(ctx, http.StatusForbidden)
			return
		case ErrUnknownPermission, sql.ErrNoRows:
			gAdapter.SetResponseStatus(ctx, http.StatusNotFound)
			return
		case ErrUnknownRole, dao.ErrCircularReference:
			gAdapter.SetResponseStatus(ctx, http.StatusInternalServerError)
			return
		}
	}

	gAdapter.SetError(ctx, v)
}

func init() {
	type Form struct {
		Name string `validator:"name,l=1-512,r=rbacname"`
		Desc string `validator:"desc,optional"`
	}

	register(
		"POST",
		"/perms",
		func(ctx context.Context) {
			ctx, tx := sqlx.Tx(ctx)
			defer tx.AutoCommit(ctx)

			MustGrantedAll(ctx, PermPermissionCreate)

			var form Form
			if err := gAdapter.ValidateForm(ctx, &form); err != nil {
				gAdapter.SetError(ctx, err)
				return
			}

			dao.NewPerm(ctx, form.Name, form.Desc)
		},
		Form{},
	)
}

func init() {
	register(
		"GET",
		"/perms",
		func(ctx context.Context) {
			MustGrantedAll(ctx, PermPermissionListAll)
			gAdapter.WriteJSON(ctx, dao.Perms(ctx))
		},
		nil,
	)
}

func init() {
	type Form struct {
		Name string `validator:"name,L=1-512,where=url,r=rbacname"`
	}

	register(
		"DELETE",
		"/perm/{name}",
		func(ctx context.Context) {
			ctx, tx := sqlx.Tx(ctx)
			defer tx.AutoCommit(ctx)

			MustGrantedAll(ctx, PermPermissionDelete)

			var form Form
			if err := gAdapter.ValidateForm(ctx, &form); err != nil {
				gAdapter.SetError(ctx, err)
				return
			}

			dao.DelPerm(ctx, form.Name)
		},
		Form{},
	)
}
