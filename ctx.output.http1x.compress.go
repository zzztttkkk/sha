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

type _ResponseBufWrapper struct {
	res *Response
}

func (w *_ResponseBufWrapper) Write(p []byte) (int, error) {
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

// br > deflate > gzip
func (ctx *RequestCtx) AutoCompress() {
	if disableCompress {
		return
	}

	acceptGzip := false
	acceptDeflate := false

	for _, headerVal := range ctx.Request.Header.GetAllRef(headerAcceptEncoding) {
		for _, v := range bytes.Split(*headerVal, headerCompressValueSep) {
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
