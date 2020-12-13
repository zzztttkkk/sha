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

var brPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressBrotli() {
	ctx.Response.Header.Set(internal.B(HeaderContentEncoding), brotliStr)
	ctx.Response.cwrPool = &brPool

	var cwr *brotli.Writer
	brI := brPool.Get()
	if brI != nil {
		cwr = brI.(*brotli.Writer)
		cwr.Reset(ctx.Response.buf)
	} else {
		cwr = brotli.NewWriterLevel(ctx.Response.buf, CompressLevelBrotli)
	}
	ctx.Response.compressWriter = cwr
}

var gzipPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressGzip() {
	ctx.Response.Header.Set(internal.B(HeaderContentEncoding), gzipStr)
	ctx.Response.cwrPool = &gzipPool

	var cwr *gzip.Writer
	var err error
	brI := gzipPool.Get()
	if brI != nil {
		cwr = brI.(*gzip.Writer)
		cwr.Reset(ctx.Response.buf)
	} else {
		cwr, err = gzip.NewWriterLevel(ctx.Response.buf, CompressLevelGzip)
		if err != nil {
			panic(err)
		}
	}
	ctx.Response.compressWriter = cwr
}

var deflatePool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressDeflate() {
	ctx.Response.Header.Set(internal.B(HeaderContentEncoding), deflateStr)
	ctx.Response.cwrPool = &deflatePool

	var cwr *flate.Writer
	var err error
	brI := deflatePool.Get()
	if brI != nil {
		cwr = brI.(*flate.Writer)
		cwr.Reset(ctx.Response.buf)
	} else {
		cwr, err = flate.NewWriter(ctx.Response.buf, CompressLevelDeflate)
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

	for _, headerVal := range ctx.Request.Header.GetAll(internal.B(HeaderAcceptEncoding)) {
		for _, v := range bytes.Split(headerVal, headerCompressValueSep) {
			switch string(v) {
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

func (res *Response) freeCompressionWriter() {
	if res.compressWriter == nil {
		return
	}
	res.compressWriter.Reset(nil)
	res.cwrPool.Put(res.compressWriter)
	res.cwrPool = nil
	res.compressWriter = nil
}
