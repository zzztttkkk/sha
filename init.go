package suna

import (
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/rbac"

	"github.com/zzztttkkk/suna/sqlu"

	"github.com/zzztttkkk/suna/secret"

	"github.com/zzztttkkk/suna/output"

	"github.com/zzztttkkk/suna/internal"

	"github.com/zzztttkkk/suna/ini"
	"github.com/zzztttkkk/suna/middleware"
)

type InitOption struct {
	IniFiles      []string
	Authenticator ctxs.Authenticator
}

var config = ini.New()

func Init(opt *InitOption) *ini.Ini {
	internal.Provide(
		func() *ini.Ini {
			for _, fn := range opt.IniFiles {
				config.Load(fn)
			}
			config.Done()
			config.Print()
			return config
		},
	)
	internal.Provide(func() ctxs.Authenticator { return opt.Authenticator })
	internal.Provide(func(conf *ini.Ini) redis.Cmdable { return config.RedisClient() })
	internal.Provide(func(conf *ini.Ini) (*sqlx.DB, []*sqlx.DB) { return config.SqlClients() })

	internal.Invoke(ctxs.Init)
	internal.Invoke(middleware.Init)
	internal.Invoke(output.Init)
	internal.Invoke(rbac.Init)
	internal.Invoke(secret.Init)
	internal.Invoke(sqlu.Init)

	return config
}
