package h

import (
	"fmt"
	"github.com/zzztttkkk/sha"
)

func NewPrintMiddleware(s string) sha.Middleware {
	return sha.NamedMiddleware(
		s,
		sha.MiddlewareFunc(
			func(ctx *sha.RequestCtx, next func()) {
				fmt.Println(s)
				next()
			},
		),
	)
}

func NewPrintHandler(s string) sha.RequestHandler {
	return sha.NamedRequestHandler(
		s,
		sha.RequestHandlerFunc(
			func(ctx *sha.RequestCtx) {
				fmt.Println(s)
			},
		),
	)
}
