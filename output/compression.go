package output

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

var (
	gzipInHeader    = []byte("gzip")
	deflateInHeader = []byte("deflate")
)

func WriteTo(ctx *fasthttp.RequestCtx, p []byte) (int, error) {
	accept := ctx.Request.Header.Peek("Accept-Encoding")
	if len(accept) < 1 {
		return ctx.Write(p)
	}
	if bytes.Contains(accept, deflateInHeader) {
		ctx.Response.Header.Set("Content-Encoding", "deflate")
		return fasthttp.WriteDeflate(ctx, p)
	}
	if bytes.Contains(accept, gzipInHeader) {
		ctx.Response.Header.Set("Content-Encoding", "gzip")
		return fasthttp.WriteGzip(ctx, p)
	}
	return ctx.Write(p)
}

type _CtxCompressionWriter struct {
	raw *fasthttp.RequestCtx
}

func (w *_CtxCompressionWriter) Write(p []byte) (int, error) {
	return WriteTo(w.raw, p)
}

// revive:disable-next-line
func NewCompressionWriter(ctx *fasthttp.RequestCtx) *_CtxCompressionWriter {
	return &_CtxCompressionWriter{raw: ctx}
}
