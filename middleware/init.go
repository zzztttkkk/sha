package middleware

import (
	"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis_rate/v8"
)

func Init(redisc redis.Cmdable) {
	rateLimiter = redis_rate.NewLimiter(redisc)
}
