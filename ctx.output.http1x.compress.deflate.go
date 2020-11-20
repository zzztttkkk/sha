package suna

import (
	"github.com/klauspost/compress/flate"
	"sync"
)

type _DeflateW struct {
	*flate.Writer
	_BytesWriter
}

func (defW *_DeflateW) setPtr(ptr *Response) {
	defW._BytesWriter.res = ptr
}

func (defW *_DeflateW) Write(p []byte) (int, error) {
	return defW.Writer.Write(p)
}

func (defW *_DeflateW) Flush() error {
	err := defW.Writer.Flush()
	defW.setPtr(nil)
	deflatePool.Put(defW)
	defW.Writer.Reset(&defW._BytesWriter)
	return err
}

var deflatePool = sync.Pool{
	New: func() interface{} {
		brw := &_DeflateW{}
		var err error
		brw.Writer, err = flate.NewWriter(&brw._BytesWriter, CompressLevelDeflate)
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
	ctx.Response.Header.Set(headerContentEncoding, deflateStr)
	ctx.Response.compressWriter = acquireDeflateW(&ctx.Response)
	ctx.Response.newCompressWriter = acquireDeflateW
}
