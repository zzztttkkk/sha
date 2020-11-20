package suna

import (
	"github.com/andybalholm/brotli"
	"sync"
)

var brPool = sync.Pool{New: func() interface{} { return nil }}

func acquireBrW(res *Response) WriteFlusher {
	v := brPool.Get().(*_CompressBr)
	if v == nil {
		v = &_CompressBr{}
		v.Writer = brotli.NewWriterLevel(res, CompressLevelBrotli)
		return v
	}
	v.Writer.Reset(res)
	return v
}

type _CompressBr struct {
	*brotli.Writer
}

func (v *_CompressBr) Write(p []byte) (int, error) {
	return v.Writer.Write(p)
}

func (v *_CompressBr) Flush() error {
	err := v.Writer.Flush()
	v.Writer.Reset(nil)
	brPool.Put(v)
	return err
}

func (ctx *RequestCtx) CompressBrotli() {
	ctx.Response.Header.Set(headerContentEncoding, brotliStr)
	ctx.Response.compressWriter = acquireBrW(&ctx.Response)
	ctx.Response.newCompressWriter = acquireBrW
}
