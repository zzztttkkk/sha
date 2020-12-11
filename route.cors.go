package sha

import (
	"github.com/zzztttkkk/sha/internal"
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

	headerKvs [][]byte
}

func (options *CorsOptions) init() {
	if options.MaxAge > 0 {
		options.headerKvs = append(options.headerKvs, internal.B(HeaderAccessControlMaxAge))
		options.headerKvs = append(options.headerKvs, []byte(strconv.FormatInt(options.MaxAge, 10)))
	}

	if len(options.AllowMethods) < 1 {
		options.AllowMethods = "GET, POST, OPTIONS"
	}

	options.headerKvs = append(options.headerKvs, internal.B(HeaderAccessControlAllowMethods))
	options.headerKvs = append(options.headerKvs, []byte(options.AllowMethods))

	if len(options.AllowHeaders) > 0 {
		options.headerKvs = append(options.headerKvs, internal.B(HeaderAccessControlAllowHeaders))
		options.headerKvs = append(options.headerKvs, []byte(options.AllowHeaders))
	}
	if options.AllowCredentials {
		options.headerKvs = append(options.headerKvs, internal.B(HeaderAccessControlAllowCredentials))
		options.headerKvs = append(options.headerKvs, []byte("true"))
	}
	if len(options.ExposeHeaders) > 0 {
		options.headerKvs = append(options.headerKvs, internal.B(HeaderAccessControlExposeHeaders))
		options.headerKvs = append(options.headerKvs, []byte(options.ExposeHeaders))
	}
	if options.OnForbidden == nil {
		options.OnForbidden = func(ctx *RequestCtx) { ctx.SetStatus(http.StatusForbidden) }
	}
}

func (options *CorsOptions) writeHeader(ctx *RequestCtx) {
	var key []byte
	for i, v := range options.headerKvs {
		if i&1 == 0 {
			key = v
		} else {
			ctx.Response.Header.Set(key, v)
		}
	}
}

func (options *CorsOptions) verify(ctx *RequestCtx) bool {
	origin, ok := ctx.Request.Header.Get(internal.B(HeaderOrigin))
	if !ok || len(origin) < 1 { // same origin
		return true
	}
	if !options.CheckOrigin(origin) {
		return false
	}
	ctx.Response.Header.Set(internal.B(HeaderAccessControlAllowOrigin), origin)
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
