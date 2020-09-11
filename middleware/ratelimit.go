package middleware

import (
	"github.com/go-redis/redis_rate/v8"
	"github.com/valyala/fasthttp"
	"time"

	"github.com/zzztttkkk/suna/router"

	"github.com/zzztttkkk/suna/output"
)

type RateLimiter struct {
	raw   *redis_rate.Limiter
	opt   *RateLimiterOption
	limit *redis_rate.Limit
}

type RateLimiterOption struct {
	Rate   int
	Period time.Duration
	Burst  int
	GetKey func(ctx *fasthttp.RequestCtx) string
}

func NewRateLimiter(opt *RateLimiterOption) *RateLimiter {
	return &RateLimiter{
		raw: redis_rate.NewLimiter(redisc),
		opt: opt,
		limit: &redis_rate.Limit{
			Rate:   opt.Rate,
			Period: opt.Period,
			Burst:  opt.Burst,
		},
	}
}

func (rl *RateLimiter) AsHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		key := rl.opt.GetKey(ctx)
		if len(key) < 1 {
			next(ctx)
			return
		}
		res, err := rl.raw.Allow(key, rl.limit)
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

func (rl *RateLimiter) AsMiddleware() fasthttp.RequestHandler {
	return rl.AsHandler(router.Next)
}
