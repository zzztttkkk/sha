package ctxs

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/rbac"

	"github.com/zzztttkkk/suna/internal"
)

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) rbac.Subject
}

type _EmptySubject struct {
}

func (subject *_EmptySubject) GetId() int64 {
	return -1
}

var authenticator Authenticator
var emptySubject rbac.Subject = &_EmptySubject{}

func Subject(ctx *fasthttp.RequestCtx) rbac.Subject {
	if authenticator == nil {
		return nil
	}

	iv, ok := ctx.UserValue(internal.RCtxKeySubject).(rbac.Subject)
	if ok {
		if iv == emptySubject {
			return nil
		}
		return iv
	}

	subject := authenticator.Auth(ctx)
	if subject == nil || subject.GetId() < 1 {
		ctx.SetUserValue(internal.RCtxKeySubject, emptySubject)
		return nil
	}

	ctx.SetUserValue(internal.RCtxKeySubject, subject)
	return subject
}
