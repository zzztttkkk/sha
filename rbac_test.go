package sha

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/rbac"
	"github.com/zzztttkkk/sha/sqlx"
	"testing"
)

type _RbacUser int64

func (u _RbacUser) GetID() int64 { return int64(u) }

func (u _RbacUser) Info(ctx context.Context) interface{} {
	if RCtx(ctx) == nil {
		return "non-request"
	}
	return RCtx(ctx).RemoteAddr().String()
}

func init() {
	sqlx.OpenWriteableDB(
		"mysql",
		"root:123456@/sha?autocommit=false",
	)
}

func Test_Rbac(t *testing.T) {
	mux := NewMux("", nil)
	server := Default(mux)
	server.Port = 8096
	mux.HandleDoc("get", "/doc")

	mux.HTTP(
		"get",
		"/redirect",
		RequestHandlerFunc(func(ctx *RequestCtx) { RedirectPermanently("https://google.com") }),
	)

	branch := NewBranch()

	auth.SetImplementation(auth.Func(func(ctx context.Context) (auth.Subject, error) { return _RbacUser(12), nil }))
	UseRBAC(branch, nil)

	rbac.GrantRoot(12)

	mux.AddBranch("/rbac", branch)

	mux.Print()
	server.ListenAndServe()
}
