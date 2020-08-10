package middleware

import (
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var redisc redis.Cmdable

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			redisc = conf.RedisClient()
		},
	)
}
