package mware

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/utils"
	"golang.org/x/sync/singleflight"
	"net/textproto"
	"time"
)

type _RedCacheT struct {
	sg          singleflight.Group
	seconds     int
	headers     map[string]bool
	statusCodes map[int]bool
}

type RedCacheOption struct {
	StatusCodes   []int
	Headers       []string
	ExpireSeconds int
}

const DisableRedCacheKey = "Snow-Disable-Redcache"

func NewRedCache(opt *RedCacheOption) *_RedCacheT {
	c := &_RedCacheT{
		seconds:     opt.ExpireSeconds,
		headers:     map[string]bool{},
		statusCodes: map[int]bool{},
	}

	if len(opt.StatusCodes) < 1 {
		opt.StatusCodes = append(opt.StatusCodes, fasthttp.StatusOK)
	}

	opt.Headers = append(opt.Headers, "Content-Type")

	for _, h := range opt.Headers {
		c.headers[textproto.CanonicalMIMEHeaderKey(h)] = true
	}

	for _, n := range opt.StatusCodes {
		c.statusCodes[n] = true
	}

	return c
}

func (c *_RedCacheT) Handler(ctx *fasthttp.RequestCtx) {
	_, _, _ = c.sg.Do(
		"do",
		func() (interface{}, error) {
			c.do(ctx)
			return nil, nil
		},
	)
}

type _ItemT struct {
	Headers []string
	Body    []byte
	Status  int
}

func (c *_RedCacheT) do(ctx *fasthttp.RequestCtx) {
	key := "snow:rcache:" + utils.B2s(ctx.Path())

	v, _ := redisClient.Get(key).Bytes()
	if len(v) > 0 {
		item := _ItemT{}
		err := json.Unmarshal(v, &item)
		if err != nil {
			redisClient.Del(key)
		} else {
			ctx.SetStatusCode(item.Status)
			var _key string
			for i, item := range item.Headers {
				if i%2 == 0 {
					_key = item
				} else {
					ctx.Response.Header.Set(_key, item)
				}
			}
			_, _ = ctx.Write(item.Body)
			return
		}
	}

	defer func() {
		if v := recover(); v != nil {
			panic(v)
		}

		if !c.statusCodes[ctx.Response.StatusCode()] {
			return
		}

		item := _ItemT{
			Headers: nil,
			Body:    ctx.Response.Body(),
			Status:  ctx.Response.StatusCode(),
		}

		disable := false

		ctx.Response.Header.VisitAll(
			func(key, value []byte) {
				if disable {
					return
				}

				skey := utils.B2s(key)
				if skey == DisableRedCacheKey {
					disable = true
					return
				}

				if c.headers[skey] {
					item.Headers = append(item.Headers, skey)
					item.Headers = append(item.Headers, utils.B2s(value))
				}
			},
		)

		bs, _ := json.Marshal(item)
		redisClient.Set(key, bs, time.Second*time.Duration(c.seconds))
	}()
	router.Next(ctx)
}
