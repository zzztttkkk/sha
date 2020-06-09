package mware

import (
	"github.com/go-redis/redis_rate/v8"
	"github.com/zzztttkkk/snow/ini"
)

func Init(reader UidReader) {
	uidReader = reader
	redisClient = ini.RedisClient()
	rateLimiter = redis_rate.NewLimiter(redisClient)
}
