package middleware

import (
	"github.com/savsgio/gotils"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/router"
)

type CorsOption struct {
	AllowOrigins     string
	MaxAge           int
	AllowMethods     string
	AllowHeaders     string
	AllowCredentials bool
	ExposeHeaders    string

	m   map[string]bool
	kvs []string
}

func (option *CorsOption) init() {
	var kvs []string
	if len(option.AllowOrigins) < 1 {
		panic("suna.middleware.cors: empty AllowOrigins")
	} else if option.AllowOrigins != "*" {
		for _, name := range strings.Split(option.AllowOrigins, ";") {
			name := strings.TrimSpace(name)
			if len(name) < 1 {
				return
			}
			option.m[name] = true
		}
	}
	if option.MaxAge > 0 {
		kvs = append(kvs, fasthttp.HeaderAccessControlMaxAge)
		kvs = append(kvs, strconv.FormatInt(int64(option.MaxAge), 10))
	}
	if len(option.AllowMethods) > 0 {
		kvs = append(kvs, fasthttp.HeaderAccessControlAllowMethods)
		kvs = append(kvs, option.AllowMethods)
	}
	if len(option.AllowHeaders) > 0 {
		kvs = append(kvs, fasthttp.HeaderAccessControlAllowHeaders)
		kvs = append(kvs, option.AllowHeaders)
	}
	if option.AllowCredentials {
		kvs = append(kvs, fasthttp.HeaderAccessControlAllowCredentials)
		kvs = append(kvs, "true")
	}
	if len(option.ExposeHeaders) > 0 {
		kvs = append(kvs, fasthttp.HeaderAccessControlExposeHeaders)
		kvs = append(kvs, option.ExposeHeaders)
	}

	option.kvs = kvs
}

func (option *CorsOption) writeHeaders(ctx *fasthttp.RequestCtx) {
	origin := ctx.Request.Header.Peek("Origin")
	if len(origin) < 1 {
		return
	}

	if len(option.m) != 0 {
		os := gotils.B2S(origin)
		_, ok := option.m[os]
		if !ok {
			return
		}
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, os)
	} else {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, "*")
	}

	var key string
	for i, v := range option.kvs {
		if i%2 == 0 {
			key = v
		} else {
			ctx.Response.Header.Set(key, v)
		}
	}
}

type _Cors struct {
	opt *CorsOption
}

func NewCors(opt *CorsOption) *_Cors {
	c := &_Cors{opt: opt}
	c.opt.init()
	return c
}

func (c *_Cors) handle(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer c.opt.writeHeaders(ctx)
		next(ctx)
	}
}

func (c *_Cors) AsMiddleware() fasthttp.RequestHandler {
	return c.handle(router.Next)
}

func (c *_Cors) OptionsHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) { c.opt.writeHeaders(ctx) }
}
