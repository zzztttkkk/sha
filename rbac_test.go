package sha

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/rbac"
	"github.com/zzztttkkk/sha/sqlx"
	"net/url"
	"testing"
)

type _RbacUser int64

func (u _RbacUser) GetID() int64 { return int64(u) }

func (u _RbacUser) Info() interface{} { return nil }

func init() {
	sqlx.OpenWriteableDB(
		"mysql",
		"root:123456@/sha?autocommit=false&parseTime=true&loc="+url.QueryEscape("Asia/Shanghai"),
	)
}

func Test_Rbac(t *testing.T) {
	mux := NewMux("", nil)
	server := Default(mux)
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

	mux.Print(false, false)
	server.ListenAndServe()
}
