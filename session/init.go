package session

import (
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var cfg *config.Type
var redisc redis.Cmdable

func init() {
	internal.LazyInvoke(
		func(conf *config.Type) {
			cfg = conf
			redisc = conf.RedisClient()

			_initSession()
		},
	)
}
