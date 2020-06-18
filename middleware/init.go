package middleware

import (
	"context"
	"github.com/go-redis/redis_rate/v8"
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/middleware/interfaces"
	"log"
)

func Init(userReader func(ctx context.Context, uid int64) interfaces.User) {
	userFetcher = userReader

	redisClient = ini.RedisClient()
	rateLimiter = redis_rate.NewLimiter(redisClient)

	authTokenInCookie = ini.GetOr("app.auth.cookiename", "")
	authTokenInHeader = ini.GetOr("app.auth.headername", "")
	if len(authTokenInHeader) < 1 && len(authTokenInCookie) < 1 {
		log.Print("!!! warning !!! snow.middleware: `app.auth.cookiename` and `app.auth.headername` are empty")
	}
}
