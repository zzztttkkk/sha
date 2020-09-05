package session

import (
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/middleware"
)

var cfg *config.Suna
var redisc redis.Cmdable
var cors *middleware.Cors

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			cfg = conf
			redisc = conf.RedisClient()

			_initSession()
		},
	)
}
