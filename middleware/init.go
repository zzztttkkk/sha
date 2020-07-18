package middleware

import (
	"github.com/go-redis/redis_rate/v8"

	"github.com/zzztttkkk/snow/ini"
)

var config *ini.Config

func Init(conf *ini.Config, autherV Auther) {
	auther = autherV
	redisClient = conf.RedisClient()
	rateLimiter = redis_rate.NewLimiter(redisClient)
	config = conf
}
