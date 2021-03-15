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
	mux := sha.NewMux(nil)
	server := sha.Default(mux)
	//mux.HandleDoc("get", "/doc")

	rbacGroup := mux.NewGroup("/rbac")

	auth.Use(ManagerFunc(func(ctx context.Context) (auth.Subject, error) {
		rctx := sha.Unwrap(ctx)
		if rctx == nil {
			return nil, sha.StatusError(sha.StatusUnauthorized)
		}
		pwd, _ := rctx.Request.Header.Get("RBAC-Password")
		name, _ := rctx.Request.Header.Get("RBAC-Name")

		if string(pwd) == "123456" && string(name) == "root-12" {
			return _RbacUser(12), nil
		}
		return nil, sha.StatusError(sha.StatusUnauthorized)
	}))
	sha.UseRBAC(rbacGroup, nil)

	rbac.GrantRoot(12)

	mux.Print()
	server.ListenAndServe()
}
