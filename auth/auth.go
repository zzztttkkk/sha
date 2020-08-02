package auth

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/rbac"
)

type User interface {
	GetId() int64
}

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) rbac.User
}
