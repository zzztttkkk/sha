package sha

import (
	"strconv"
)

type CorsOption struct {
	AllowMethods     string `json:"allow_methods" toml:"allow-methods"`
	AllowHeaders     string `json:"allow_headers" toml:"allow-headers"`
	ExposeHeaders    string `json:"expose_headers" toml:"expose-headers"`
	AllowCredentials bool   `json:"allow_credentials" toml:"allow-credentials"`
	MaxAge           int64  `json:"max_age" toml:"max-age"`
}

type CorsOptions struct {
	headerKeys []string
	headerVals [][]byte
}

type CORSOriginChecker func(origin []byte) *CorsOptions

func NewCorsOptions(cc *CorsOption) *CorsOptions {
	v := &CorsOptions{}

	if cc.MaxAge > 0 {
		v.headerKeys = append(v.headerKeys, HeaderAccessControlMaxAge)
		v.headerVals = append(v.headerVals, []byte(strconv.FormatInt(cc.MaxAge, 10)))
	}

	if len(cc.AllowMethods) < 1 {
		cc.AllowMethods = "GET, POST, OPTIONS"
	}

	v.headerKeys = append(v.headerKeys, HeaderAccessControlAllowMethods)
	v.headerVals = append(v.headerVals, []byte(cc.AllowMethods))

	if len(cc.AllowHeaders) > 0 {
		v.headerKeys = append(v.headerKeys, HeaderAccessControlAllowHeaders)
		v.headerVals = append(v.headerVals, []byte(cc.AllowHeaders))
	}
	if cc.AllowCredentials {
		v.headerKeys = append(v.headerKeys, HeaderAccessControlAllowCredentials)
		v.headerVals = append(v.headerVals, []byte("true"))
	}
	if len(cc.ExposeHeaders) > 0 {
		v.headerKeys = append(v.headerKeys, HeaderAccessControlExposeHeaders)
		v.headerVals = append(v.headerVals, []byte(cc.ExposeHeaders))
	}
	return v
}

func (options *CorsOptions) writeHeader(ctx *RequestCtx, origin []byte) {
	header := &ctx.Response.Header
	header.Set(HeaderAccessControlAllowOrigin, origin)
	for i := 0; i < len(options.headerKeys); i++ {
		header.Append(options.headerKeys[i], options.headerVals[i])
	}
}
