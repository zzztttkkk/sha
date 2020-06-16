package mware

import (
	"encoding/hex"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/utils"
	"hash"
)

func NewSigner(hash hash.Hash, secretGen func(ctx *fasthttp.RequestCtx) []byte) fasthttp.RequestHandler {
	var pool = utils.NewBytesPool(hash.Size()*2, hash.Size()*2)

	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			hash.Write(ctx.Response.Body())
			hash.Write(secretGen(ctx))
			result := pool.Get()
			hex.Encode(*result, hash.Sum(nil))
			defer pool.Put(result)

			ctx.Response.Header.SetBytesKV([]byte("Body-Sign"), *result)
		}()
		router.Next(ctx)
	}
}
