package auth

import (
	"context"
	"errors"
	"github.com/zzztttkkk/suna/rbac/internal"
)

type Subject interface {
	GetID() int64
	Info() interface{}
}

type Authenticator interface {
	Auth(ctx context.Context) (Subject, bool)
}

var authenticator Authenticator

func Auth(ctx context.Context) (Subject, bool) {
	return authenticator.Auth(ctx)
}

type Func func(ctx context.Context) (Subject, bool)

func (fn Func) Auth(ctx context.Context) (Subject, bool) {
	return fn(ctx)
}

var ErrUnauthenticatedOperation = errors.New("suna.rbac: unauthenticated operation")

func MustAuth(ctx context.Context) Subject {
	s, o := Auth(ctx)
	if !o {
		panic(ErrUnauthenticatedOperation)
	}
	return s
}

func init() {
	internal.Dig.Provide(
		func(v Authenticator) internal.AuthOK {
			authenticator = v
			return 0
		},
	)
}
