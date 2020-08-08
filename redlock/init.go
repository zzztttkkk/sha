package redlock

import (
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var redisc redis.Cmdable

func init() {
	internal.LazyInvoke(
		func(conf *config.Config) {
			redisc = conf.RedisClient()
		},
	)
}
