package suna

import (
	"github.com/klauspost/compress/gzip"
	"sync"
)

type _GzipW struct {
	*gzip.Writer
	_ResponseBufWrapper
}

func (gW *_GzipW) setPtr(ptr *Response) {
	gW._ResponseBufWrapper.res = ptr
}

func (gW *_GzipW) Write(p []byte) (int, error) {
	return gW.Writer.Write(p)
}

func (gW *_GzipW) Flush() error {
	err := gW.Writer.Flush()
	gW.setPtr(nil)
	gzipWPool.Put(gW)
	gW.Writer.Reset(&gW._ResponseBufWrapper)
	return err
}

var gzipWPool = sync.Pool{
	New: func() interface{} {
		gw := &_GzipW{}
		var err error
		gw.Writer, err = gzip.NewWriterLevel(&gw._ResponseBufWrapper, CompressLevelGzip)
		if err != nil {
			panic(err)
		}
		return gw
	},
}

func acquireGzipW(res *Response) WriteFlusher {
	v := gzipWPool.Get().(*_GzipW)
	v.setPtr(res)
	v.Writer.Reset(&v._ResponseBufWrapper)
	return v
}

func (ctx *RequestCtx) CompressGzip() {
	ctx.Response.Header.Set(headerContentEncoding, gzipStr)
	ctx.Response.compressWriter = acquireGzipW(&ctx.Response)
	ctx.Response.newCompressWriter = acquireGzipW
}
