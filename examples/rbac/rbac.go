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

func (u _RbacUser) Info(ctx context.Context) interface{} { return nil }

func init() {
	sqlx.OpenWriteableDB("mysql", "root:123456@/sha?autocommit=false")
}

type ManagerFunc func(ctx context.Context) (auth.Subject, error)

func (f ManagerFunc) Auth(ctx context.Context) (auth.Subject, error) { return f(ctx) }

func main() {
	mux := sha.NewMux(nil, nil)
	server := sha.Default(mux)
	mux.HandleDoc("get", "/doc")

	mux.HTTPWithOptions(
		"get",
		"/redirect",
		sha.RequestHandlerFunc(func(ctx *sha.RequestCtx) { sha.RedirectPermanently("https://google.com") }),
	)

	branch := sha.NewBranch()

	auth.Use(ManagerFunc(func(ctx context.Context) (auth.Subject, error) {
		ctx = rbac.UnwrapRequestCtx(ctx)
		if ctx == nil {
			return nil, sha.StatusError(sha.StatusUnauthorized)
		}

		rctx := ctx.(*sha.RequestCtx)
		pwd, _ := rctx.Request.Header.Get("RBAC-Password")
		name, _ := rctx.Request.Header.Get("RBAC-Name")

		if string(pwd) == "123456" && string(name) == "root-12" {
			return _RbacUser(12), nil
		}
		return nil, sha.StatusError(sha.StatusUnauthorized)
	}))
	sha.UseRBAC(branch, nil)

	rbac.GrantRoot(12)

	mux.AddBranch("/rbac", branch)

	mux.Print()
	server.ListenAndServe()
}
