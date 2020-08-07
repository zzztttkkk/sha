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

func init() {
	internal.LazyInvoke(
		func(aor Authenticator) { authenticator = aor },
	)
}

func GetUser(ctx *fasthttp.RequestCtx) User {
	return authenticator.Auth(ctx)
}

func GetUserMust(ctx *fasthttp.RequestCtx) (u User) {
	if u := authenticator.Auth(ctx); u == nil {
		panic(output.HttpErrors[fasthttp.StatusUnauthorized])
	}
	return
}
