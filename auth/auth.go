package auth

import (
	"context"
	"errors"
	"github.com/zzztttkkk/sha/internal"
	"net/http"
	"sync"
)

type Subject interface {
	GetID() int64
	Info(ctx context.Context) interface{}
}

type Manager interface {
	Auth(ctx context.Context) (Subject, error)
}

var implOnce sync.Once
var impl Manager

func Auth(ctx context.Context) (Subject, error) { return impl.Auth(ctx) }

func MustAuth(ctx context.Context) Subject {
	s, o := Auth(ctx)
	if o != nil {
		panic(o)
	}
	return s
}

func Init(manager Manager) { implOnce.Do(func() { impl = manager }) }

var ErrUnauthenticatedOperation = errors.New("sha.rbac: unauthenticated operation")

func init() { internal.ErrorStatusByValue[ErrUnauthenticatedOperation] = http.StatusUnauthorized }
