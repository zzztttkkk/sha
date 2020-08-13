package auth

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/output"
)

type User interface {
	GetId() int64
}

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) User
}

var authenticator Authenticator

func GetUser(ctx *fasthttp.RequestCtx) User {
	ui := ctx.UserValue(internal.RCtxUserKey)
	if ui != nil {
		return ui.(User)
	}
	u := authenticator.Auth(ctx)
	ctx.SetUserValue(internal.RCtxUserKey, u)
	return u
}

func MustGetUser(ctx *fasthttp.RequestCtx) (u User) {
	if u = GetUser(ctx); u == nil {
		panic(output.HttpErrors[fasthttp.StatusUnauthorized])
	}
	return
}
