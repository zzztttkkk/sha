package middleware

import (
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/router"

	"github.com/zzztttkkk/suna/utils"
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
		os := utils.B2s(origin)
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

func (option *CorsOption) BindOptions(path string, router *router.Router) {
	option.init()
	router.OPTIONS(path, option.writeHeaders)
}

func NewCorsMiddleware(option *CorsOption) fasthttp.RequestHandler {
	return NewCorsHandler(option, router.Next)
}

func NewCorsHandler(option *CorsOption, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	option.init()
	return func(ctx *fasthttp.RequestCtx) {
		defer option.writeHeaders(ctx)
		next(ctx)
	}
}
