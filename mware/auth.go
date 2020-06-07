package mware

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/router"
	"unsafe"
)

type User interface {
	Identify() int64
	Pointer() unsafe.Pointer
}

type UserReader interface {
	Read(ctx *fasthttp.RequestCtx) (User, error)
}

var author UserReader

const userKey = "/u"

func GetUser(ctx *fasthttp.RequestCtx) (User, error) {
	user, ok := ctx.UserValue(userKey).(User)
	if ok {
		return user, nil
	}
	user, err := author.Read(ctx)
	if err != nil {
		return nil, err
	}
	if user != nil {
		ctx.SetUserValue(userKey, user)
	}
	return user, nil
}

func GetUserMust(ctx *fasthttp.RequestCtx) User {
	user, err := GetUser(ctx)
	if err != nil || user == nil {
		panic(output.StdErrors[fasthttp.StatusBadRequest])
	}
	return user
}

func AuthHandler(ctx *fasthttp.RequestCtx) {
	GetUserMust(ctx)
	router.Next(ctx)
}
