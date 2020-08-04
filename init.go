package suna

import (
	"context"
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/sqls"
	"sort"

	"github.com/zzztttkkk/suna/internal"

	"github.com/zzztttkkk/suna/middleware"
)

type InitOption struct {
	ConfigFiles   []string
	Authenticator auth.Authenticator
}

var cfg *config.Type

func Init(opt *InitOption) *config.Type {
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
	internal.Provide(func() auth.Authenticator { return opt.Authenticator })
	internal.Provide(func(conf *config.Type) redis.Cmdable { return cfg.RedisClient() })
	internal.Provide(
		func() *internal.RbacDi {
			return &internal.RbacDi{
				WrapCtx:         ctxs.Std,
				GetUserFromRCtx: ctxs.User,
				GetUserFromCtx: func(ctx context.Context) auth.User {
					return ctxs.User(ctxs.RequestCtx(ctx))
				},
			}
		},
	)

	internal.Invoke(ctxs.Init)
	internal.Invoke(middleware.Init)
	internal.Invoke(output.Init)
	internal.Invoke(rbac.Init)
	internal.Invoke(secret.Init)
	internal.Invoke(sqls.Init)

	return cfg
}
