package ctxs

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/utils"
)

const LoginTokenInHeader = "X-Blog-Auth"

func GetUid(ctx *fasthttp.RequestCtx) int64 {
	token := ctx.Request.Header.Peek(LoginTokenInHeader)
	if token == nil {
		return -1
	}

	clm, e := secret.JwtDecode(utils.B2s(token))
	if e != nil {
		return -1
	}
	m, ok := clm.(jwt.MapClaims)
	if !ok {
		return -1
	}
	return m["uid"].(int64)
}
