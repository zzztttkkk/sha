package sha

import (
	"strconv"
)

type CorsOptions struct {
	Name             string `json:"name" toml:"name"`
	AllowMethods     string `json:"allow_methods" toml:"allow-methods"`
	AllowHeaders     string `json:"allow_headers" toml:"allow-headers"`
	ExposeHeaders    string `json:"expose_headers" toml:"expose-headers"`
	AllowCredentials bool   `json:"allow_credentials" toml:"allow-credentials"`
	MaxAge           int64  `json:"max_age" toml:"max-age"`
}

type _CorsOptions struct {
	headerKeys []string
	headerVals [][]byte
}

type OriginToName func(origin []byte) string

var allowAllCorsOption *_CorsOptions

func init() {
	allowAllCorsOption = newCorsOptions(
		&CorsOptions{
			Name: "*", AllowMethods: "*",
			AllowHeaders: "*", ExposeHeaders: "*",
			AllowCredentials: true,
		},
	)
}

func newCorsOptions(cc *CorsOptions) *_CorsOptions {
	v := &_CorsOptions{}

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

func (options *_CorsOptions) writeHeader(ctx *RequestCtx, origin []byte) {
	header := ctx.Response.Header()
	header.Set(HeaderAccessControlAllowOrigin, origin)
	for i := 0; i < len(options.headerKeys); i++ {
		header.Append(options.headerKeys[i], options.headerVals[i])
	}
}
