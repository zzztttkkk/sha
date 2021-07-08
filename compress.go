package sha

import (
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"strings"
	"sync"
)

const (
	CompressionTypeGzip    = "gzip"
	CompressionTypeDeflate = "deflate"
	CompressionTypeBrotli  = "br"
)

var (
	CompressionLevelGzip    = gzip.DefaultCompression
	CompressionLevelDeflate = flate.DefaultCompression
	CompressionLevelBrotli  = brotli.DefaultCompression
)

type _CompressionWriter interface {
	io.Writer
	Flush() error
	Reset(writer io.Writer)
}

var brWriterPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressBrotli() {
	ctx.Response.Header().SetString(HeaderContentEncoding, CompressionTypeBrotli)
	var cwr *brotli.Writer
	brI := brWriterPool.Get()
	w := &ctx.Response._HTTPPocket

	if brI != nil {
		cwr = brI.(*brotli.Writer)
		cwr.Reset(w)
	} else {
		cwr = brotli.NewWriterLevel(w, CompressionLevelBrotli)
	}
	ctx.Response.cw = cwr
	ctx.Response.cwPool = &brWriterPool
}

var gzipWriterPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressGzip() {
	ctx.Response.Header().SetString(HeaderContentEncoding, CompressionTypeGzip)

	var cwr *gzip.Writer
	var err error
	w := &ctx.Response._HTTPPocket

	brI := gzipWriterPool.Get()
	if brI != nil {
		cwr = brI.(*gzip.Writer)
		cwr.Reset(w)
	} else {
		cwr, err = gzip.NewWriterLevel(w, CompressionLevelGzip)
		if err != nil {
			panic(err)
		}
	}
	ctx.Response.cw = cwr
	ctx.Response.cwPool = &gzipWriterPool
}

var deflateWriterPool = sync.Pool{New: func() interface{} { return nil }}

func (ctx *RequestCtx) CompressDeflate() {
	ctx.Response.Header().SetString(HeaderContentEncoding, CompressionTypeDeflate)

	var cwr *flate.Writer
	var err error
	w := &ctx.Response._HTTPPocket
	brI := deflateWriterPool.Get()
	if brI != nil {
		cwr = brI.(*flate.Writer)
		cwr.Reset(w)
	} else {
		cwr, err = flate.NewWriter(w, CompressionLevelDeflate)
		if err != nil {
			panic(err)
		}
	}
	ctx.Response.cw = cwr
	ctx.Response.cwPool = &deflateWriterPool
}

var disableCompress = false

// DisableCompress keep raw response body when debugging
func DisableCompress() { disableCompress = true }

// AutoCompress br > deflate > gzip
func (ctx *RequestCtx) AutoCompress() {
	if disableCompress {
		return
	}

	if ctx.Response.cw != nil {
		return
	}

	for _, headerVal := range ctx.Request.Header().GetAll(HeaderAcceptEncoding) {
		hsv := utils.S(headerVal)
		if strings.Contains(hsv, CompressionTypeBrotli) {
			ctx.CompressBrotli()
			return
		}
		if strings.Contains(hsv, CompressionTypeDeflate) {
			ctx.CompressDeflate()
			return
		}
		if strings.Contains(hsv, CompressionTypeGzip) {
			ctx.CompressGzip()
			return
		}
	}
}
