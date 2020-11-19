package h

import (
	"fmt"
	"github.com/zzztttkkk/suna"
)

func NewPrintMiddleware(s string) suna.Middleware {
	return suna.MiddlewareFunc(
		func(ctx *suna.RequestCtx, next func()) {
			fmt.Println(s)
			next()
		},
	)
}

func NewPrintHandler(s string) suna.RequestHandler {
	return suna.RequestHandlerFunc(
		func(ctx *suna.RequestCtx) {
			fmt.Println(s)
		},
	)
}
