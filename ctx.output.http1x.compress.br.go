package suna

import (
	"github.com/andybalholm/brotli"
	"github.com/zzztttkkk/suna/internal"
	"sync"
)

type _CompressBrotli struct {
	*brotli.Writer
	_ResponseBufWrapper
}

func (brw *_CompressBrotli) setPtr(ptr *Response) {
	brw._ResponseBufWrapper.buf = ptr.buf
}

func (brw *_CompressBrotli) Write(p []byte) (int, error) {
	return brw.Writer.Write(p)
}

func (brw *_CompressBrotli) Flush() error {
	err := brw.Writer.Flush()
	brw.setPtr(nil)
	brPool.Put(brw)
	brw.Writer.Reset(&brw._ResponseBufWrapper)
	return err
}

var brPool = sync.Pool{
	New: func() interface{} {
		brw := &_CompressBrotli{}
		brw.Writer = brotli.NewWriterLevel(&brw._ResponseBufWrapper, CompressLevelBrotli)
		return brw
	},
}

func acquireBrW(res *Response) WriteFlusher {
	v := brPool.Get().(*_CompressBrotli)
	v.setPtr(res)
	v.Writer.Reset(&v._ResponseBufWrapper)
	return v
}

func (ctx *RequestCtx) CompressBrotli() {
	ctx.Response.Header.Set(internal.B(HeaderContentEncoding), brotliStr)
	ctx.Response.compressWriter = acquireBrW(&ctx.Response)
	ctx.Response.newCompressWriter = acquireBrW
}
