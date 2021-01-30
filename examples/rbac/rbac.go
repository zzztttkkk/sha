package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zzztttkkk/sha"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/rbac"
	"github.com/zzztttkkk/sha/sqlx"
)

type _RbacUser int64

func (u _RbacUser) GetID() int64 { return int64(u) }

func (u _RbacUser) Info(ctx context.Context) interface{} {
	return sha.MustToRCtx(ctx).RemoteAddr().String()
}

func init() {
	sqlx.OpenWriteableDB(
		"mysql",
		"root:123456@/sha?autocommit=false",
	)
}

func main() {
	mux := sha.NewMux(nil, nil)
	server := sha.Default(mux)
	mux.HandleDoc("get", "/doc")

	mux.HTTP(
		"get",
		"/redirect",
		sha.RequestHandlerFunc(func(ctx *sha.RequestCtx) { sha.RedirectPermanently("https://google.com") }),
	)

	branch := sha.NewBranch()

	auth.SetImplementation(auth.Func(func(ctx context.Context) (auth.Subject, error) { return _RbacUser(12), nil }))
	sha.UseRBAC(branch, nil)

	rbac.GrantRoot(12)

	mux.AddBranch("/rbac", branch)

	mux.Print()
	server.ListenAndServe()
}
