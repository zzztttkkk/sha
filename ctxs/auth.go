package ctxs

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"

	"github.com/zzztttkkk/suna/internal"
)

type _EmptyUser struct {
}

func (subject *_EmptyUser) GetId() int64 {
	return -1
}

var authenticator auth.Authenticator
var emptyUser auth.User = &_EmptyUser{}

func User(ctx *fasthttp.RequestCtx) auth.User {
	if authenticator == nil {
		return nil
	}

	iv, ok := ctx.UserValue(internal.RCtxKeyUser).(auth.User)
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
