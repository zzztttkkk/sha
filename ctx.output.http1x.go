package suna

import (
	"github.com/zzztttkkk/suna/internal"
	"net/http"
	"strconv"
)

var responseHeaderBufferPool = internal.NewBufferPoll(4096)
var statusTextMap = map[int][]byte{}

func init() {
	for i := 0; i < 600; i++ {
		txt := http.StatusText(i)
		if len(txt) < 1 {
			continue
		}
		statusTextMap[i] = []byte(txt)
	}
}

var newline = []byte("\r\n")
var headerKVSep = []byte(": ")

func (ctx *RequestCtx) sendHttp1xResponseBuffer() error {
	if ctx.compressW != nil {
		_ = ctx.compressW.Close()
	}

	res := &ctx.Response
	res.Header.SetContentLength(int64(len(res.buf)))
	if !res.headerWritten {
		err := ctx.writeHttp1xHeader()
		if err != nil {
			return err
		}
	}
	if _, err := ctx.conn.Write(newline); err != nil {
		return err
	}

	_, err := ctx.conn.Write(ctx.Response.buf)
	return err
}

var ErrNilContentLength = newStdError("suna: nil content length")
var ErrUnknownSResponseStatusCode = newStdError("suna: unknown response status code")

func (ctx *RequestCtx) writeHttp1xHeader() error {
	res := &ctx.Response
	res.headerWritten = true

	_, exists := res.Header.Get(contentLengthKey)
	if !exists {
		return ErrNilContentLength
	}

	if res.statusCode < 1 {
		res.statusCode = 200
	}

	statusTxt := statusTextMap[res.statusCode]
	if len(statusTxt) < 1 {
		return ErrUnknownSResponseStatusCode
	}

	buf := responseHeaderBufferPool.Get()
	defer responseHeaderBufferPool.Put(buf)

	buf.Data = append(buf.Data, ctx.protocol.Version...)
	buf.Data = append(buf.Data, ' ')
	buf.Data = append(buf.Data, internal.B(strconv.FormatInt(int64(res.statusCode), 10))...)
	buf.Data = append(buf.Data, ' ')

	buf.Data = append(buf.Data, statusTxt...)
	buf.Data = append(buf.Data, '\r', '\n')

	ctx.Response.Header.EachItem(
		func(k, v []byte) bool {
			buf.Data = append(buf.Data, k...)
			buf.Data = append(buf.Data, headerKVSep...)
			buf.Data = append(buf.Data)
			quoteHeaderValueToBuf(v, &(buf.Data))
			buf.Data = append(buf.Data, newline...)
			return true
		},
	)

	_, err := ctx.conn.Write(buf.Data)
	return err
}

func (ctx *RequestCtx) Write(p []byte) (int, error) {
	res := &ctx.Response
	if res.compressW != nil {
		return res.compressW.Write(p)
	}
	return res.Write(p)
}

func (ctx *RequestCtx) WriteString(s string) (int, error) {
	return ctx.Write(internal.B(s))
}

func (ctx *RequestCtx) WriteError(err HttpError) {
	res := &ctx.Response
	res.Reset()

	res.statusCode = err.StatusCode()
	if header := err.Header(); header != nil {
		header.EachItem(
			func(k, v []byte) bool {
				res.Header.Append(k, v)
				return true
			},
		)
	}
	res.buf = append(res.buf, err.Message()...)
}
