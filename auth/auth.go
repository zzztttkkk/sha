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
	return 0
}

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) (User, bool)
}

var authenticator Authenticator
var emptyUser = &_EmptyUser{}

func GetUser(ctx *fasthttp.RequestCtx) (User, bool) {
	ui := ctx.UserValue(internal.RCtxUserKey)
	if ui != nil {
		if ui == emptyUser {
			return nil, false
		}
		return ui.(User), true
	}

	u, ok := authenticator.Auth(ctx)
	if ok {
		ctx.SetUserValue(internal.RCtxUserKey, u)
		return u, true
	}

	ctx.SetUserValue(internal.RCtxUserKey, emptyUser)
	return nil, false
}

func MustGetUser(ctx *fasthttp.RequestCtx) User {
	v, ok := GetUser(ctx)
	if ok {
		return v
	}
	panic(output.HttpErrors[fasthttp.StatusUnauthorized])
}

func Reset(ctx *fasthttp.RequestCtx) {
	ctx.SetUserValue(internal.RCtxUserKey, nil)
}
