package suna

import (
	"context"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/cache"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/reflectx"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/session"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
	"log"
	"reflect"
	"sort"
	"strings"
	"sync"
)

var rkKeyWarnOnce = sync.Once{}

func doReservedKeyWarning() {
	rkKeyWarnOnce.Do(
		func() {
			log.Printf("suna: reserved suna.session.Session keys: `%s`,", internal.SessionExistsKey)
		},
	)
}

type InitOption struct {
	ConfigFiles   []string
	Authenticator auth.Authenticator
}

var cfg *config.Type

func Init(opt *InitOption) *config.Type {
	doReservedKeyWarning()

	internal.Provide(
		func() *config.Type {
			if opt == nil || len(opt.ConfigFiles) < 1 {
				cfg = config.New()
				return cfg
			}
			sort.Sort(sort.StringSlice(opt.ConfigFiles))
			cfg = config.FromFiles(opt.ConfigFiles...)
			return cfg
		},
	)
	internal.Provide(
		func() auth.Authenticator {
			if opt == nil {
				return nil
			}
			return opt.Authenticator
		},
	)

	internal.Provide(
		func() *internal.RbacDi {
			return &internal.RbacDi{
				WrapCtx:        ctxs.Wrap,
				GetUserFromCtx: func(ctx context.Context) auth.User { return auth.GetUserMust(ctxs.Unwrap(ctx)) },
			}
		},
	)

	_LoadSubModules()
	internal.Invoke()
	return cfg
}

// trigger internal.LazyInvoke
func _LoadSubModules() string {
	buf := strings.Builder{}

	buf.WriteString(reflect.ValueOf(cache.NewLru).String())
	buf.WriteString(reflect.ValueOf(ctxs.Unwrap).String())
	buf.WriteString(reflect.ValueOf(middleware.NewAccessLogger).String())
	buf.WriteString(reflect.ValueOf(output.Error).String())
	buf.WriteString(reflect.ValueOf(rbac.Loader).String())
	buf.WriteString(reflect.ValueOf(reflectx.ExportedKeys).String())
	buf.WriteString(reflect.ValueOf(secret.AesDecrypt).String())
	buf.WriteString(reflect.ValueOf(session.Get).String())
	buf.WriteString(reflect.ValueOf(sqls.CreateTable).String())
	buf.WriteString(reflect.ValueOf(validator.RegisterFunc).String())

	return buf.String()
}
