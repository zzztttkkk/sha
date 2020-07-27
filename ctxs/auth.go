package ctxs

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/rbac"

	"github.com/zzztttkkk/suna/internal"
)

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) rbac.User
}

type _EmptyUser struct {
}

func (subject *_EmptyUser) GetId() int64 {
	return -1
}

var authenticator Authenticator
var emptyUser rbac.User = &_EmptyUser{}

func User(ctx *fasthttp.RequestCtx) rbac.User {
	if authenticator == nil {
		return nil
	}

	iv, ok := ctx.UserValue(internal.RCtxKeyUser).(rbac.User)
	if ok {
		if iv == emptyUser {
			return nil
		}
		return iv
	}

	subject := authenticator.Auth(ctx)
	if subject == nil || subject.GetId() < 1 {
		ctx.SetUserValue(internal.RCtxKeyUser, emptyUser)
		return nil
	}

	ctx.SetUserValue(internal.RCtxKeyUser, subject)
	return subject
}
