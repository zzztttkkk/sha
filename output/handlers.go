package output

import (
	"github.com/valyala/fasthttp"
)

func Recover(ctx *fasthttp.RequestCtx, val interface{}) {
	if val == nil {
		return
	}
	switch v := val.(type) {
	case error:
		Error(ctx, v)
	default:
		Error(ctx, HttpErrors[fasthttp.StatusInternalServerError])
	}
}

func NotFound(ctx *fasthttp.RequestCtx) {
	Error(ctx, HttpErrors[fasthttp.StatusNotFound])
}

func MethodNotAllowed(ctx *fasthttp.RequestCtx) {
	Error(ctx, HttpErrors[fasthttp.StatusMethodNotAllowed])
}
