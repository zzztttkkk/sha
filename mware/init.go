package mware

import (
	"github.com/go-redis/redis_rate/v8"
	"github.com/zzztttkkk/snow/redisc"
)

func Init(authenticator Authenticator) {
	author = authenticator
	redisClient = redisc.Client()
	rateLimiter = redis_rate.NewLimiter(redisClient)
}
