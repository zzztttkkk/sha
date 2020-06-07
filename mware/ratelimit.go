package mware

import (
	"github.com/go-redis/redis_rate/v8"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
	"time"
)

var rateLimiter *redis_rate.Limiter

func NewRateLimiter(keyFunc func(*fasthttp.RequestCtx) string, duration time.Duration, rate int) fasthttp.RequestHandler {
	var option = &redis_rate.Limit{
		Rate:   rate,
		Period: duration,
		Burst:  rate,
	}

	return func(ctx *fasthttp.RequestCtx) {
		key := keyFunc(ctx)
		if len(key) < 1 {
			router.Next(ctx)
			return
		}

		res, err := rateLimiter.Allow(key, option)
		if err != nil {
			output.Error(ctx, err)
			return
		}
		if !res.Allowed {
			output.StdError(ctx, fasthttp.StatusTooManyRequests)
			return
		}
		router.Next(ctx)
	}
}
