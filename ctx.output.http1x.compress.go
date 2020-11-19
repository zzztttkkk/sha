package suna

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"io"
)

var (
	headerContentEncoding  = []byte("Content-Encoding")
	headerAcceptEncoding   = []byte("Accept-Encoding")
	headerCompressValueSep = []byte(", ")
	gzipStr                = []byte("gzip")
	deflateStr             = []byte("deflate")
	brotliStr              = []byte("br")

	CompressLevelGzip    = gzip.DefaultCompression
	CompressLevelDeflate = flate.DefaultCompression
	CompressLevelBrotli  = brotli.DefaultCompression
)

type WriteFlusher interface {
	io.Writer
	Flush() error
}

type _BytesWriter struct {
	res *Response
}

func (w *_BytesWriter) Write(p []byte) (int, error) {
	if w.res == nil {
		return 0, nil
	}
	w.res.buf = append(w.res.buf, p...)
	return len(p), nil
}

var disableCompress = false

// DisableCompress keep raw response body when debugging
func DisableCompress() {
	disableCompress = true
}

func (ctx *RequestCtx) AutoCompress() {
	if disableCompress {
		return
	}

	for _, headerVal := range ctx.Request.Header.GetAllRef(headerAcceptEncoding) {
		for _, v := range bytes.Split(*headerVal, headerCompressValueSep) {
			switch string(v) {
			case "gzip":
				ctx.CompressGzip()
				return
			case "deflate":
				ctx.CompressDeflate()
				return
			case "br":
				ctx.CompressBrotli()
				return
			}
		}
	}
}
