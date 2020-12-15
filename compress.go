package sha

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/zzztttkkk/sha/internal"
	"io"
	"sync"
)

var (
	headerCompressValueSep = []byte(", ")
	gzipStr                = []byte("gzip")
	deflateStr             = []byte("deflate")
	brotliStr              = []byte("br")

	CompressLevelGzip    = gzip.DefaultCompression
	CompressLevelDeflate = flate.DefaultCompression
	CompressLevelBrotli  = brotli.DefaultCompression
)

type _CompressionWriter interface {
	io.Writer
	Flush() error
	Reset(writer io.Writer)
}

var brWriterPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressBrotli() {
	ctx.Response.Header.Set(HeaderContentEncoding, brotliStr)
	ctx.Response.compressWriterPool = &brWriterPool

	var cwr *brotli.Writer
	brI := brWriterPool.Get()
	if brI != nil {
		cwr = brI.(*brotli.Writer)
		cwr.Reset(ctx.Response.bodyBuf)
	} else {
		cwr = brotli.NewWriterLevel(ctx.Response.bodyBuf, CompressLevelBrotli)
	}
	ctx.Response.compressWriter = cwr
}

var gzipWriterPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressGzip() {
	ctx.Response.Header.Set(HeaderContentEncoding, gzipStr)
	ctx.Response.compressWriterPool = &gzipWriterPool

	var cwr *gzip.Writer
	var err error
	brI := gzipWriterPool.Get()
	if brI != nil {
		cwr = brI.(*gzip.Writer)
		cwr.Reset(ctx.Response.bodyBuf)
	} else {
		cwr, err = gzip.NewWriterLevel(ctx.Response.bodyBuf, CompressLevelGzip)
		if err != nil {
			panic(err)
		}
	}
	ctx.Response.compressWriter = cwr
}

var deflateWriterPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressDeflate() {
	ctx.Response.Header.Set(HeaderContentEncoding, deflateStr)
	ctx.Response.compressWriterPool = &deflateWriterPool

	var cwr *flate.Writer
	var err error
	brI := deflateWriterPool.Get()
	if brI != nil {
		cwr = brI.(*flate.Writer)
		cwr.Reset(ctx.Response.bodyBuf)
	} else {
		cwr, err = flate.NewWriter(ctx.Response.bodyBuf, CompressLevelDeflate)
		if err != nil {
			panic(err)
		}
	}
	ctx.Response.compressWriter = cwr
}

var disableCompress = false

// DisableCompress keep raw response body when debugging
func DisableCompress() {
	disableCompress = true
}

// br > deflate > gzip
func (ctx *RequestCtx) AutoCompress() {
	if disableCompress {
		return
	}

	acceptGzip := false
	acceptDeflate := false

	for _, headerVal := range ctx.Request.Header.GetAll(HeaderAcceptEncoding) {
		for _, v := range bytes.Split(headerVal, headerCompressValueSep) {
			switch internal.S(v) {
			case "gzip":
				acceptGzip = true
			case "deflate":
				acceptDeflate = true
			case "br":
				ctx.CompressBrotli()
				return
			}
		}
	}

	if acceptDeflate {
		ctx.CompressDeflate()
		return
	}

	if acceptGzip {
		ctx.CompressGzip()
	}
}

func (res *Response) freeWriter() {
	res.sendBuf = nil

	if res.compressWriter == nil {
		return
	}
	res.compressWriter.Reset(nil)
	res.compressWriterPool.Put(res.compressWriter)
	res.compressWriterPool = nil
	res.compressWriter = nil
}
