package rbac

import (
	"context"
	"database/sql"
	"github.com/zzztttkkk/suna/rbac/auth"
	"github.com/zzztttkkk/suna/rbac/dao"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/sqlx"
	"net/http"
)

func register(method, path string, fn HandlerFunc, form interface{}) {
	internal.Dig.Append(
		func(router Router, _ _PermOK) {
			router.HandleWithDoc(method, path, fn, form)
		},
	)
}

func Recover(rw ReqWriter) {
	v := recover()
	if v == nil {
		return
	}

	switch tv := v.(type) {
	case error:
		switch tv {
		case auth.ErrUnauthenticatedOperation:
			rw.SetStatus(http.StatusUnauthorized)
			return
		case ErrPermissionDenied:
			rw.SetStatus(http.StatusForbidden)
			return
		case ErrUnknownPermission, sql.ErrNoRows:
			rw.SetStatus(http.StatusNotFound)
			return
		case ErrUnknownRole, dao.ErrCircularReference:
			rw.SetStatus(http.StatusInternalServerError)
			return
		}
	}

	panic(v)
}

func init() {
	type Form struct {
		Name string `validate:"name,L=1-512"`
		Desc string `validate:"desc,optional"`
	}

	register(
		"POST",
		"/perms",
		func(rctx context.Context, rw ReqWriter) {
			ctx, committer := sqlx.Tx(rctx)
			defer committer()

			MustGrantedAll(ctx, PermPermissionCreate)

			var form Form
			rw.MustValidate(&form)
			dao.NewPerm(ctx, form.Name, form.Desc)
		},
		Form{},
	)
}

func init() {
	register(
		"GET",
		"/perms",
		func(rctx context.Context, rw ReqWriter) {
			MustGrantedAll(rctx, PermPermissionListAll)
			ret := dao.Perms(rctx)
			rw.WriteJSON(ret)
		},
		nil,
	)
}

func init() {
	type Form struct {
		Name string `validate:",L=1-512,P=name"`
	}

	register(
		"DELETE",
		"/perm/:name",
		func(rctx context.Context, rw ReqWriter) {
			ctx, committer := sqlx.Tx(rctx)
			defer committer()

			MustGrantedAll(ctx, PermPermissionDelete)

			var form Form
			rw.MustValidate(&form)
			dao.DelPerm(ctx, form.Name)
		},
		Form{},
	)
}
