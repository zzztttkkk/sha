package auth

import (
	"context"
	"errors"
	"github.com/zzztttkkk/sha/internal"
	"net/http"
)

type Subject interface {
	GetID() int64
	Info() interface{}
}

type Interface interface {
	Auth(ctx context.Context) (Subject, error)
}

var impl Interface

func Auth(ctx context.Context) (Subject, error) {
	return impl.Auth(ctx)
}

type Func func(ctx context.Context) (Subject, error)

func (fn Func) Auth(ctx context.Context) (Subject, error) {
	return fn(ctx)
}

var ErrUnauthenticatedOperation = errors.New("sha.rbac: unauthenticated operation")

func init() {
	internal.ErrorStatusByValue[ErrUnauthenticatedOperation] = http.StatusUnauthorized
}

func MustAuth(ctx context.Context) Subject {
	s, o := Auth(ctx)
	if o != nil {
		panic(o)
	}
	return s
}

func SetImplementation(authenticator Interface) {
	impl = authenticator
}
