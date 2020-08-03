package auth

import (
	"github.com/valyala/fasthttp"
)

type User interface {
	GetId() int64
}

type Authenticator interface {
	Auth(*fasthttp.RequestCtx) User
}
