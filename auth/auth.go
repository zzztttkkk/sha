package auth

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/output"
)

type User interface {
	GetId() int64
}

type _EmptyUser struct {
}

func (*_EmptyUser) GetId() int64 {
	return -1
}

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) (User, bool)
}

type AuthenticatorFunc func(*fasthttp.RequestCtx) (User, bool)

func (fn AuthenticatorFunc) Auth(ctx *fasthttp.RequestCtx) (User, bool) {
	return fn(ctx)
}

var _Authenticator Authenticator
var emptyUser = &_EmptyUser{}

// Use the global `Authenticator` to get the user information of the current request
// and cache it in the `fasthttp.RequestCtx.UserValue` (even if the authentication fails).
func GetUser(ctx *fasthttp.RequestCtx) (User, bool) {
	ui := ctx.UserValue(internal.RCtxUserKey)
	if ui != nil {
		if ui == emptyUser {
			return nil, false
		}
		return ui.(User), true
	}

	u, ok := _Authenticator.Auth(ctx)
	if ok {
		ctx.SetUserValue(internal.RCtxUserKey, u)
		return u, true
	}

	ctx.SetUserValue(internal.RCtxUserKey, emptyUser)
	return nil, false
}

// If authentication fails, a 401 exception will be thrown.
func MustGetUser(ctx *fasthttp.RequestCtx) User {
	v, ok := GetUser(ctx)
	if ok {
		return v
	}
	panic(output.HttpErrors[fasthttp.StatusUnauthorized])
}

// Clear the user info cache
func Reset(ctx *fasthttp.RequestCtx) {
	ctx.SetUserValue(internal.RCtxUserKey, nil)
}

func IsAvailable() bool {
	return _Authenticator != nil
}
