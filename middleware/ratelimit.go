package middleware

import (
	"github.com/go-redis/redis_rate/v8"
	"github.com/valyala/fasthttp"
	"time"

	"github.com/zzztttkkk/suna/output"
)

type RateLimiter struct {
	raw       *redis_rate.Limiter
	limit     *redis_rate.Limit
	keyGetter func(ctx *fasthttp.RequestCtx) string
}

func NewRateLimiter(rate int, period time.Duration, burst int, keyGetter func(ctx *fasthttp.RequestCtx) string) *RateLimiter {
	return &RateLimiter{
		raw: redis_rate.NewLimiter(redisc),
		limit: &redis_rate.Limit{
			Rate:   rate,
			Period: period,
			Burst:  burst,
		},
		keyGetter: keyGetter,
	}
}

func (rl *RateLimiter) Process(ctx *fasthttp.RequestCtx, next func()) {
	key := rl.keyGetter(ctx)
	if len(key) < 1 {
		next()
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
	next()
}
