package suna

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/sqlx"
	"net/url"
	"testing"
	"time"
)

type _RbacUser int64

func (u _RbacUser) GetID() int64 { return int64(u) }

func (u _RbacUser) Info() interface{} { return "rbac.TestUser" }

func init() {
	sqlx.OpenWriteableDB(
		"mysql",
		"root:123456@/suna_test?autocommit=false&parseTime=true&loc="+url.QueryEscape("Asia/Shanghai"),
	)
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

	auth.SetImplementation(auth.Func(func(ctx context.Context) (auth.Subject, error) { return _RbacUser(12), nil }))
	UseRBAC(branch, "", nil, false)

	rbac.GrantRoot(12)

	mux.AddBranch("/rbac", branch)

	go func() {
		time.Sleep(time.Second)
		mux.Print(false, false)
	}()

	server.ListenAndServe()
}
