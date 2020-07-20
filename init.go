package suna

import (
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"

	"github.com/zzztttkkk/suna/sqls"

	"github.com/zzztttkkk/suna/secret"

	"github.com/zzztttkkk/suna/output"

	"github.com/zzztttkkk/suna/internal"

	"github.com/zzztttkkk/suna/ini"
	"github.com/zzztttkkk/suna/middleware"
)

type InitOption struct {
	IniFiles []string
	Auther   middleware.Auther
}

var config = ini.New()

func Init(opt *InitOption) *ini.Config {
	internal.Provide(
		func() *ini.Config {
			for _, fn := range opt.IniFiles {
				config.Load(fn)
			}
			config.Done()
			config.Print()
			return config
		},
	)
	internal.Provide(func() middleware.Auther { return opt.Auther })
	internal.Provide(func(conf *ini.Config) redis.Cmdable { return config.RedisClient() })
	internal.Provide(func(conf *ini.Config) (*sqlx.DB, []*sqlx.DB) { return config.SqlClients() })

	internal.Invoke(middleware.Init)
	internal.Invoke(output.Init)
	internal.Invoke(secret.Init)
	internal.Invoke(sqls.Init)

	return config
}
