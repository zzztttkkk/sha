package suna

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/rbac/auth"
	"github.com/zzztttkkk/suna/sqlx"
	"net/url"
	"testing"
)

type _RbacUser struct {
	id int64
}

func (u *_RbacUser) GetID() int64 { return u.id }

func init() {
	sqlx.OpenWriteableDB(
		"mysql",
		"root:123456@/suna_test?autocommit=false&parseTime=true&loc="+url.QueryEscape("Asia/Shanghai"),
	)
	sqlx.EnableLogging()
}

func Test_Rbac(t *testing.T) {
	mux := NewMux("", nil)
	server := Default(mux)
	mux.HandleDoc("get", "/doc")

	branch := NewBranch()
	branch.Use(
		MiddlewareFunc(
			func(ctx *RequestCtx, next func()) {
				defer rbac.Recover(ctx)
				next()
			},
		),
	)

	UseRBAC(
		branch,
		auth.Func(func(ctx context.Context) (auth.Subject, bool) { return &_RbacUser{id: 12}, true }),
		"",
		false,
	)

	rbac.GrantRoot(12)

	mux.AddBranch("/rbac", branch)

	server.ListenAndServe()
}
