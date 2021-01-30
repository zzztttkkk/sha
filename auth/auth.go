package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/zzztttkkk/sha/internal"
)

type Subject interface {
	GetID() int64
	Info(ctx context.Context) interface{}
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
