package suna

import (
	"github.com/klauspost/compress/flate"
	"github.com/zzztttkkk/suna/internal"
	"sync"
)

type _DeflateW struct {
	*flate.Writer
	_ResponseBufWrapper
}

func (defW *_DeflateW) setPtr(ptr *Response) {
	defW._ResponseBufWrapper.res = ptr
}

func (defW *_DeflateW) Write(p []byte) (int, error) {
	return defW.Writer.Write(p)
}

func (defW *_DeflateW) Flush() error {
	err := defW.Writer.Flush()
	defW.setPtr(nil)
	deflatePool.Put(defW)
	defW.Writer.Reset(&defW._ResponseBufWrapper)
	return err
}

var deflatePool = sync.Pool{
	New: func() interface{} {
		brw := &_DeflateW{}
		var err error
		brw.Writer, err = flate.NewWriter(&brw._ResponseBufWrapper, CompressLevelDeflate)
		if err != nil {
			panic(err)
		}
		return brw
	},
}

func acquireDeflateW(res *Response) WriteFlusher {
	v := deflatePool.Get().(*_DeflateW)
	v.setPtr(res)
	return v
}

func (ctx *RequestCtx) CompressDeflate() {
	ctx.Response.Header.Set(internal.B(HeaderContentEncoding), deflateStr)
	ctx.Response.compressWriter = acquireDeflateW(&ctx.Response)
	ctx.Response.newCompressWriter = acquireDeflateW
}
