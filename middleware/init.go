package middleware

import (
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var redisc redis.Cmdable
var cors_option *config.Cors

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			redisc = conf.RedisClient()
			cors_option = &conf.Cors
			initCors()
		},
	)
}
