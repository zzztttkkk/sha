package mware

import (
	"context"
	"github.com/go-redis/redis_rate/v8"
	"github.com/zzztttkkk/snow/ini"
	"log"
)

func Init(userReader func(ctx context.Context, uid int64) User) {
	readUserById = userReader

	redisClient = ini.RedisClient()
	rateLimiter = redis_rate.NewLimiter(redisClient)

	authTokenInCookie = ini.GetOr("app.auth.cookiename", "")
	authTokenInHeader = ini.GetOr("app.auth.headername", "")
	if len(authTokenInHeader) < 1 && len(authTokenInCookie) < 1 {
		log.Print("!!! warning !!! snow.mware: `app.auth.cookiename` and `app.auth.headername` are empty")
	}
}
