package output

import (
	"github.com/valyala/fasthttp"
)

func Recover(ctx *fasthttp.RequestCtx, val interface{}) {
	switch v := val.(type) {
	case error:
		Error(ctx, v)
	default:
		Error(ctx, HttpErrors[fasthttp.StatusInternalServerError])
	}
}
