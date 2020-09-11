package output

import (
	"log"

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

func RecoverAndLogging(ctx *fasthttp.RequestCtx, val interface{}) {
	if val == nil {
		return
	}
	switch v := val.(type) {
	case error:
		Error(ctx, v)
	default:
		Error(ctx, HttpErrors[fasthttp.StatusInternalServerError])
	}

	if ctx.Response.StatusCode() > 499 {
		log.Printf("suna.output.recover: %v", val)
	}
}

func NotFound(ctx *fasthttp.RequestCtx) {
	Error(ctx, HttpErrors[fasthttp.StatusNotFound])
}

func MethodNotAllowed(ctx *fasthttp.RequestCtx) {
	Error(ctx, HttpErrors[fasthttp.StatusMethodNotAllowed])
}
