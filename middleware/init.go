package middleware

import (
	"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis_rate/v8"

	"github.com/zzztttkkk/suna/ini"
)

var config *ini.Config

func Init(conf *ini.Config, autherV Auther, redisc redis.Cmdable) {
	auther = autherV
	redisClient = redisc
	rateLimiter = redis_rate.NewLimiter(redisClient)
	config = conf
}
