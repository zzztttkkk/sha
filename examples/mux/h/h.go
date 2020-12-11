package h

import (
	"fmt"
	"github.com/zzztttkkk/sha"
)

func NewPrintMiddleware(s string) sha.Middleware {
	return sha.MiddlewareFunc(
		func(ctx *sha.RequestCtx, next func()) {
			fmt.Println(s)
			next()
		},
	)
}

func NewPrintHandler(s string) sha.RequestHandler {
	return sha.RequestHandlerFunc(
		func(ctx *sha.RequestCtx) {
			fmt.Println(s)
		},
	)
}
