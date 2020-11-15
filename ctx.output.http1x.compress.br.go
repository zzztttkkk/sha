package suna

import (
	"github.com/andybalholm/brotli"
	"sync"
)

type _BrW struct {
	*brotli.Writer
	_BytesWriter
}

func (brw *_BrW) setPtr(ptr *Response) {
	brw._BytesWriter.res = ptr
}

func (brw *_BrW) Write(p []byte) (int, error) {
	return brw.Writer.Write(p)
}

func (brw *_BrW) Flush() error {
	err := brw.Writer.Flush()
	brw.setPtr(nil)
	brPool.Put(brw)
	brw.Writer.Reset(&brw._BytesWriter)
	return err
}

var brPool = sync.Pool{
	New: func() interface{} {
		brw := &_BrW{}
		brw.Writer = brotli.NewWriterLevel(&brw._BytesWriter, CompressLevelBrotli)
		return brw
	},
}

func acquireBrW(res *Response) WriteFlusher {
	v := brPool.Get().(*_BrW)
	v.setPtr(res)
	v.Writer.Reset(&v._BytesWriter)
	return v
}

func (ctx *RequestCtx) CompressBrotli() {
	ctx.Response.Header.Set(headerContentEncoding, brotliStr)
	ctx.Response.compressW = acquireBrW(&ctx.Response)
	ctx.Response.compressFunc = acquireBrW
}
