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
	brotliStr              = []byte("brotli")

	CompressLevelGzip    = gzip.DefaultCompression
	CompressLevelDeflate = flate.DefaultCompression
	CompressLevelBrotli  = brotli.DefaultCompression
)

type _BytesWriter struct {
	ptr *[]byte
}

func (w *_BytesWriter) Write(p []byte) (int, error) {
	*w.ptr = append(*w.ptr, p...)
	return len(p), nil
}

func (ctx *RequestCtx) CompressGzip() {
	var cW io.WriteCloser
	var err error
	cW, err = gzip.NewWriterLevel(&_BytesWriter{ptr: &ctx.Response.buf}, CompressLevelGzip)
	if err != nil {
		panic(err)
	}

	ctx.Response.Header.Set(headerContentEncoding, gzipStr)
	ctx.compressW = cW
	ctx.Response.compressW = cW
}

func (ctx *RequestCtx) CompressDeflate() {
	var cW io.WriteCloser
	var err error
	cW, err = flate.NewWriter(&_BytesWriter{ptr: &ctx.Response.buf}, CompressLevelDeflate)
	if err != nil {
		panic(err)
	}
	ctx.Response.Header.Set(headerContentEncoding, deflateStr)
	ctx.compressW = cW
	ctx.Response.compressW = cW
}

func (ctx *RequestCtx) CompressBrotli() {
	var cW io.WriteCloser
	cW = brotli.NewWriterLevel(&_BytesWriter{ptr: &ctx.Response.buf}, CompressLevelBrotli)
	ctx.Response.Header.Set(headerContentEncoding, brotliStr)
	ctx.compressW = cW
	ctx.Response.compressW = cW
}

func (ctx *RequestCtx) AutoCompress() {
	headerVal, ok := ctx.Request.Header.Get(headerAcceptEncoding)
	if !ok || len(headerVal) < 1 {
		return
	}

	for _, v := range bytes.Split(headerVal, headerCompressValueSep) {
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
