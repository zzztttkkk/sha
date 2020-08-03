package internal

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
)

type RbacDi struct {
	WrapCtx         func(ctx *fasthttp.RequestCtx) context.Context
	GetUserFromRCtx func(ctx *fasthttp.RequestCtx) auth.User
	GetUserFromCtx  func(ctx context.Context) auth.User
}
