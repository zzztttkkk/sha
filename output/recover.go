package output

import (
	"github.com/go-errors/errors"
	"github.com/valyala/fasthttp"
	"log"
)

func Recover(ctx *fasthttp.RequestCtx, val interface{}) {
	switch v := val.(type) {
	case error:
		Error(ctx, v)
	default:
		log.Println(errors.Wrap(v, 1).ErrorStack())
		StdError(ctx, fasthttp.StatusInternalServerError)
	}
}
