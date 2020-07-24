package middleware

import (
	"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis_rate/v8"

	"github.com/zzztttkkk/suna/ini"
)

var config *ini.Ini

func Init(conf *ini.Ini, redisc redis.Cmdable) {
	rateLimiter = redis_rate.NewLimiter(redisc)
	config = conf
}
