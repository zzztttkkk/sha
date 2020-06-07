package mware

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
	"unsafe"
)

type User interface {
	Id() int64
	Pointer() unsafe.Pointer
}

type Authenticator interface {
	Auth(ctx *fasthttp.RequestCtx) (User, error)
}

var author Authenticator

const userKey = "/u"

func GetUser(ctx *fasthttp.RequestCtx) User {
	user, ok := ctx.UserValue(userKey).(User)
	if ok {
		return user
	}
	user, err := author.Auth(ctx)
	if err != nil {
		panic(err)
	}
	if user != nil {
		ctx.SetUserValue(userKey, user)
	}
	return user
}

func GetUserMust(ctx *fasthttp.RequestCtx) User {
	user := GetUser(ctx)
	if user == nil {
		panic(output.NewHttpError(fasthttp.StatusUnauthorized, -1, ""))
	}
	return user
}

func AuthHandler(ctx *fasthttp.RequestCtx) {
	GetUserMust(ctx)
	router.Next(ctx)
}
