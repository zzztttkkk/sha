package suna

import (
	"net/http"
	"strconv"
)

type CorsOptions struct {
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int64
	CheckOrigin      func(origin []byte) bool
	OnForbidden      func(ctx *RequestCtx)

	headerKvs [][]byte
}

var (
	headerAccessControlAllowCredentials = []byte("Access-Control-Allow-Credentials")
	headerAccessControlAllowHeaders     = []byte("Access-Control-Allow-Headers")
	headerAccessControlAllowMethods     = []byte("Access-Control-Allow-Methods")
	headerAccessControlExposeHeaders    = []byte("Access-Control-Expose-Headers")
	headerAccessControlMaxAge           = []byte("Access-Control-Max-Age")
	headerAccessControlAllowOrigin      = []byte("Access-Control-Allow-Origin")
	headerOrigin                        = []byte("Origin")
)

func (options *CorsOptions) init() {
	if options.MaxAge > 0 {
		options.headerKvs = append(options.headerKvs, headerAccessControlMaxAge)
		options.headerKvs = append(options.headerKvs, []byte(strconv.FormatInt(options.MaxAge, 10)))
	}
	if len(options.AllowMethods) > 0 {
		options.headerKvs = append(options.headerKvs, headerAccessControlAllowMethods)
		options.headerKvs = append(options.headerKvs, []byte(options.AllowMethods))
	}
	if len(options.AllowHeaders) > 0 {
		options.headerKvs = append(options.headerKvs, headerAccessControlAllowHeaders)
		options.headerKvs = append(options.headerKvs, []byte(options.AllowHeaders))
	}
	if options.AllowCredentials {
		options.headerKvs = append(options.headerKvs, headerAccessControlAllowCredentials)
		options.headerKvs = append(options.headerKvs, []byte("true"))
	}
	if len(options.ExposeHeaders) > 0 {
		options.headerKvs = append(options.headerKvs, headerAccessControlExposeHeaders)
		options.headerKvs = append(options.headerKvs, []byte(options.ExposeHeaders))
	}
	if options.OnForbidden == nil {
		options.OnForbidden = func(ctx *RequestCtx) { ctx.WriteStatus(http.StatusForbidden) }
	}
}

func (options *CorsOptions) writeHeader(ctx *RequestCtx) {
	origin, ok := ctx.Request.Header.Get(headerOrigin)
	if len(origin) < 1 || !ok { // same origin
		return
	}
	var key []byte
	for i, v := range options.headerKvs {
		if i%2 == 0 {
			key = v
		} else {
			ctx.Response.Header.Set(key, v)
		}
	}
}

func (options *CorsOptions) verify(ctx *RequestCtx) bool {
	origin, ok := ctx.Request.Header.Get(headerOrigin)
	if !ok || len(origin) < 1 { // same origin
		return true
	}
	if !options.CheckOrigin(origin) {
		return false
	}
	ctx.Response.Header.Set(headerAccessControlAllowOrigin, origin)
	return true
}

func (options *CorsOptions) Process(ctx *RequestCtx, next func()) {
	if !options.verify(ctx) {
		options.OnForbidden(ctx)
		return
	}
	defer options.writeHeader(ctx)
	next()
}
