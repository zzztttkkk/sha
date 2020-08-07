package middleware

import (
	"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis_rate/v8"
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/router"

	"github.com/zzztttkkk/suna/output"
)

var rateLimiter *redis_rate.Limiter

type _RateLimiter struct {
	raw *redis_rate.Limiter
	opt *RateLimiterOption
}

type RateLimiterOption struct {
	redis_rate.Limit
	GetKey func(ctx *fasthttp.RequestCtx) string
}

func NewRateLimiter(redisC redis.Cmdable, opt *RateLimiterOption) *_RateLimiter {
	return &_RateLimiter{
		raw: redis_rate.NewLimiter(redisC),
		opt: opt,
	}
}

func (rl *_RateLimiter) AsHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		key := rl.opt.GetKey(ctx)
		if len(key) < 1 {
			next(ctx)
			return
		}
		res, err := rateLimiter.Allow(key, &(rl.opt.Limit))
		if err != nil {
			output.Error(ctx, err)
			return
		}
		if !res.Allowed {
			output.Error(ctx, output.HttpErrors[fasthttp.StatusTooManyRequests])
			return
		}
		next(ctx)
	}
}

func (rl *_RateLimiter) AsMiddleware() fasthttp.RequestHandler {
	return rl.AsHandler(router.Next)
}
