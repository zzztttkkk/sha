package sha

import (
	"strconv"
)

type CorsOptions struct {
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int64

	headerKeys []string
	headerVals [][]byte
	inited     bool
}

func (options *CorsOptions) Init() *CorsOptions {
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
	return options
}

func (options *CorsOptions) writeHeader(ctx *RequestCtx, origin []byte) {
	header := &ctx.Response.Header
	header.Set(HeaderAccessControlAllowOrigin, origin)
	for i := 0; i < len(options.headerKeys); i++ {
		header.Append(options.headerKeys[i], options.headerVals[i])
	}
}
