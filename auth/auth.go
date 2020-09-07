package auth

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/output"
)

// User is an interface, can `GetId() int64`
type User interface {
	GetId() int64
}

// Authenticator is an interface, do auth for `fasthttp.RequestCtx`
type Authenticator interface {
	Auth(*fasthttp.RequestCtx) (User, bool)
}

// AuthenticatorFunc auth function
type AuthenticatorFunc func(*fasthttp.RequestCtx) (User, bool)

// Auth do auth
func (fn AuthenticatorFunc) Auth(ctx *fasthttp.RequestCtx) (User, bool) {
	return fn(ctx)
}

var authenticatorV Authenticator

// GetUser Use the global `Authenticator` to get the user information of the current request
// and cache it in the `fasthttp.RequestCtx.UserValue` (even if the authentication fails).
func GetUser(ctx *fasthttp.RequestCtx) (User, bool) {
	ui := ctx.UserValue(internal.RCtxUserKey)
	if ui != nil {
		switch rv := ui.(type) {
		case User:
			return rv, true
		default:
			return nil, false
		}
	}

	u, ok := authenticatorV.Auth(ctx)
	if ok {
		ctx.SetUserValue(internal.RCtxUserKey, u)
		return u, true
	}

	ctx.SetUserValue(internal.RCtxUserKey, 0)
	return nil, false
}

// MustGetUser If authentication fails, a 401 exception will be thrown.
func MustGetUser(ctx *fasthttp.RequestCtx) User {
	v, ok := GetUser(ctx)
	if ok {
		return v
	}
	panic(output.HttpErrors[fasthttp.StatusUnauthorized])
}

// Reset Clear the user info cache
func Reset(ctx *fasthttp.RequestCtx) {
	ctx.SetUserValue(internal.RCtxUserKey, nil)
}

// IsAvailable check the global `Authenticator` is not nil
func IsAvailable() bool {
	return authenticatorV != nil
}
