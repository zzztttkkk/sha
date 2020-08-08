package session

import (
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var cfg *config.Config
var redisc redis.Cmdable

func init() {
	internal.LazyInvoke(
		func(conf *config.Config) {
			cfg = conf
			redisc = conf.RedisClient()

			_initSession()
		},
	)
}
