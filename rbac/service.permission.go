package rbac

import (
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
			router.HandleWithDoc(method, path, fn, validator.NewMarkdownDocument(form, validator.Undefined))
		},
	)
}

func Recover(rctx RCtx) {
	v := recover()
	if v == nil {
		return
	}

	switch tv := v.(type) {
	case error:
		switch tv {
		case auth.ErrUnauthenticatedOperation:
			rctx.SetStatus(http.StatusUnauthorized)
			return
		case ErrPermissionDenied:
			rctx.SetStatus(http.StatusForbidden)
			return
		case ErrUnknownPermission, sql.ErrNoRows:
			rctx.SetStatus(http.StatusNotFound)
			return
		case ErrUnknownRole, dao.ErrCircularReference:
			rctx.SetStatus(http.StatusInternalServerError)
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
		func(rctx RCtx) {
			ctx, committer := sqlx.Tx(wrapCtx(rctx))
			defer committer()

			MustGrantedAll(ctx, PermPermissionCreate)

			var form Form
			rctx.MustValidate(&form)
			dao.NewPerm(ctx, form.Name, form.Desc)
		},
		Form{},
	)
}

func init() {
	register(
		"GET",
		"/perms",
		func(rctx RCtx) {
			MustGrantedAll(rctx, PermPermissionListAll)
			ret := dao.Perms(rctx)
			rctx.WriteJSON(ret)
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
		func(rctx RCtx) {
			ctx, committer := sqlx.Tx(wrapCtx(rctx))
			defer committer()

			MustGrantedAll(ctx, PermPermissionDelete)

			var form Form
			rctx.MustValidate(&form)
			dao.DelPerm(ctx, form.Name)
		},
		Form{},
	)
}
