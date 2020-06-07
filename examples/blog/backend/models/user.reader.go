package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/internal"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"unsafe"
)

func (user *User) Identify() int64 {
	return user.Id
}

func (user *User) Pointer() unsafe.Pointer {
	return unsafe.Pointer(user)
}

type _UserReaderT func(ctx *fasthttp.RequestCtx) (mware.User, error)

func (fn _UserReaderT) Read(ctx *fasthttp.RequestCtx) (mware.User, error) {
	return fn(ctx)
}

const LoginTokenInHeader = "X-Blog-Read"

var UserReader = _UserReaderT(
	func(ctx *fasthttp.RequestCtx) (mware.User, error) {
		token := ctx.Request.Header.Peek(LoginTokenInHeader)
		if token == nil {
			return nil, output.StdErrors[fasthttp.StatusUnauthorized]
		}

		clm, e := secret.JwtDecode(internal.B2s(token))
		if e != nil {
			return nil, output.StdErrors[fasthttp.StatusBadRequest]
		}
		m, ok := clm.(jwt.MapClaims)
		if !ok {
			return nil, output.StdErrors[fasthttp.StatusBadRequest]
		}
		user := &User{}
		user.Id = m["uid"].(int64)
		return user, nil
	},
)
