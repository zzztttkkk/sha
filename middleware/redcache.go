package middleware

import (
	"encoding/json"
	"net/textproto"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/sync/singleflight"

	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/utils"
)

type _RedCacheT struct {
	sg          singleflight.Group
	seconds     int
	headers     map[string]bool
	statusCodes map[int]bool
	getKey      func(ctx *fasthttp.RequestCtx) string
}

type RedCacheOption struct {
	StatusCodes   []int
	Headers       []string
	ExpireSeconds int
	GetKey        func(ctx *fasthttp.RequestCtx) string
}

const DisableRedCacheKey = "Snow-Disable-Redcache"

func NewRedCache(opt *RedCacheOption) *_RedCacheT {
	c := &_RedCacheT{
		seconds:     opt.ExpireSeconds,
		headers:     map[string]bool{},
		statusCodes: map[int]bool{},
		getKey:      opt.GetKey,
	}

	if c.getKey == nil {
		c.getKey = func(ctx *fasthttp.RequestCtx) string { return utils.B2s(ctx.Path()) }
	}

	if len(opt.StatusCodes) < 1 {
		opt.StatusCodes = append(opt.StatusCodes, fasthttp.StatusOK)
	}

	opt.Headers = append(opt.Headers, "Content-Type")
	opt.Headers = append(opt.Headers, "Content-Encoding")

	for _, h := range opt.Headers {
		c.headers[textproto.CanonicalMIMEHeaderKey(h)] = true
	}

	for _, n := range opt.StatusCodes {
		c.statusCodes[n] = true
	}

	return c
}

type _ItemT struct {
	Headers []string
	Body    []byte
	Status  int
}

var itemPool = sync.Pool{New: func() interface{} { return &_ItemT{} }}

func acquireItem() *_ItemT {
	return itemPool.Get().(*_ItemT)
}

func releaseItem(item *_ItemT) {
	item.Body = item.Body[:0]
	item.Headers = item.Headers[:0]
	itemPool.Put(item)
}

func (c *_RedCacheT) AsMiddleware() fasthttp.RequestHandler {
	return c.AsHandler(router.Next)
}

func (c *_RedCacheT) AsHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	item := acquireItem()
	defer releaseItem(item)

	return func(ctx *fasthttp.RequestCtx) {
		cacheKey := c.getKey(ctx)
		if len(cacheKey) < 1 {
			next(ctx)
			return
		}

		_, _, _ = c.sg.Do(
			cacheKey,
			func() (interface{}, error) {
				c.loadItem(ctx, next, item)
				return nil, nil
			},
		)

		ctx.SetStatusCode(item.Status)
		ctx.SetBody(item.Body)

		var key string
		for ind, item := range item.Headers {
			if ind%2 == 0 {
				key = item
			} else {
				ctx.Response.Header.Set(key, item)
			}
		}
	}
}

func (c *_RedCacheT) loadItem(ctx *fasthttp.RequestCtx, handler fasthttp.RequestHandler, item *_ItemT) {
	key := "snow:rcache:" + c.getKey(ctx)

	v, _ := redisClient.Get(key).Bytes()
	if len(v) > 0 {
		err := json.Unmarshal(v, item)
		if err == nil {
			return
		}
		redisClient.Del(key)
	}

	defer func() {
		if v := recover(); v != nil {
			panic(v)
		}

		disable := false

		item.Body = ctx.Response.Body();
		item.Status = ctx.Response.StatusCode()
		ctx.Response.Header.VisitAll(
			func(key, value []byte) {
				if disable {
					return
				}

				skey := utils.B2s(key)
				if skey == DisableRedCacheKey {
					disable = true
				}

				if c.headers[skey] {
					item.Headers = append(item.Headers, skey)
					item.Headers = append(item.Headers, utils.B2s(value))
				}
			},
		)

		if !c.statusCodes[ctx.Response.StatusCode()] || disable {
			return
		}

		bs, _ := json.Marshal(item)
		redisClient.Set(key, bs, time.Second*time.Duration(c.seconds))
	}()

	handler(ctx)
}
