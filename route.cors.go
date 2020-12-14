package sha

import (
	"net/http"
	"strconv"
)

type CorsOptions struct {
	AllowMethods     string                   `json:"allow_methods" toml:"allow-methods"`
	AllowHeaders     string                   `json:"allow_headers" toml:"allow-headers"`
	ExposeHeaders    string                   `json:"expose_headers" toml:"expose-headers"`
	AllowCredentials bool                     `json:"allow_credentials" toml:"allow-credentials"`
	MaxAge           int64                    `json:"max_age" toml:"max-age"`
	CheckOrigin      func(origin []byte) bool `json:"-" toml:"-"`
	OnForbidden      func(ctx *RequestCtx)    `json:"-" toml:"-"`

	headerKeys []string
	headerVals [][]byte
}

func (options *CorsOptions) init() {
	if options.MaxAge > 0 {
		options.headerKeys = append(options.headerKeys, HeaderAccessControlMaxAge)
		options.headerVals = append(options.headerVals, []byte(strconv.FormatInt(options.MaxAge, 10)))
	}

	if len(options.AllowMethods) < 1 {
		options.AllowMethods = "GET, POST, OPTIONS"
	}

	options.headerKeys = append(options.headerKeys, HeaderAccessControlAllowMethods)
	options.headerVals = append(options.headerVals, []byte(options.AllowMethods))

	if len(options.AllowHeaders) > 0 {
		options.headerKeys = append(options.headerKeys, HeaderAccessControlAllowHeaders)
		options.headerVals = append(options.headerVals, []byte(options.AllowHeaders))
	}
	if options.AllowCredentials {
		options.headerKeys = append(options.headerKeys, HeaderAccessControlAllowCredentials)
		options.headerVals = append(options.headerVals, []byte("true"))
	}
	if len(options.ExposeHeaders) > 0 {
		options.headerKeys = append(options.headerKeys, HeaderAccessControlExposeHeaders)
		options.headerVals = append(options.headerVals, []byte(options.ExposeHeaders))
	}
	if options.OnForbidden == nil {
		options.OnForbidden = func(ctx *RequestCtx) { ctx.SetStatus(http.StatusForbidden) }
	}
}

func (options *CorsOptions) writeHeader(ctx *RequestCtx) {
	header := &ctx.Response.Header
	for i := 0; i < len(options.headerKeys); i++ {
		header.Append(options.headerKeys[i], options.headerVals[i])
	}
}

func (options *CorsOptions) verify(ctx *RequestCtx) bool {
	origin, ok := ctx.Request.Header.Get(HeaderOrigin)
	if !ok || len(origin) < 1 { // same origin
		return true
	}
	if !options.CheckOrigin(origin) {
		return false
	}
	ctx.Response.Header.Set(HeaderAccessControlAllowOrigin, origin)
	return true
}

func (options *CorsOptions) Process(ctx *RequestCtx, next func()) {
	if !options.verify(ctx) {
		options.OnForbidden(ctx)
		return
	}
	options.writeHeader(ctx)
	next()
}
