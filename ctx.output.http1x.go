package suna

import (
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"net/http"
	"strconv"
)

var responseHeaderBufferPool = internal.NewBufferPoll(4096)

var newline = []byte("\r\n")
var headerKVSep = []byte(": ")

func (ctx *RequestCtx) sendHttp1xResponseBuffer() error {
	res := &ctx.Response
	if res.compressWriter != nil {
		_ = res.compressWriter.Flush()
		res.compressWriter = nil
	}

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

var ErrNilContentLength = fmt.Errorf("suna: nil content length")
var ErrUnknownResponseStatusCode = fmt.Errorf("suna: ")

func (ctx *RequestCtx) writeHttp1xHeader() error {
	res := &ctx.Response
	res.headerWritten = true

	_, exists := res.Header.Get(headerContentLength)
	if !exists {
		return ErrNilContentLength
	}

	if res.statusCode < 1 {
		res.statusCode = 200
	}

	statusTxt := statusTextMap[res.statusCode]
	if len(statusTxt) < 1 {
		return ErrUnknownResponseStatusCode
	}

	buf := responseHeaderBufferPool.Get()
	defer responseHeaderBufferPool.Put(buf)

	buf.Data = append(buf.Data, ctx.protocol.Version...)
	buf.Data = append(buf.Data, ' ')
	buf.Data = append(buf.Data, internal.B(strconv.FormatInt(int64(res.statusCode), 10))...)
	buf.Data = append(buf.Data, ' ')

	buf.Data = append(buf.Data, statusTxt...)
	buf.Data = append(buf.Data, '\r', '\n')

	res.Header.EachItem(
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
	return ctx.Response.Write(p)
}

func (ctx *RequestCtx) WriteString(s string) (int, error) {
	return ctx.Write(internal.B(s))
}

func (ctx *RequestCtx) WriteError(err error) {
	res := &ctx.Response
	res.ResetBodyBuffer()

	switch rv := err.(type) {
	case HttpResponseError:
		res.statusCode = rv.StatusCode()
		header := rv.Header()
		if header != nil {
			header.EachItem(
				func(k, v []byte) bool {
					res.Header.Append(k, v)
					return true
				},
			)
		}
		_, _ = res.Write(rv.Body())
	case HttpError:
		res.statusCode = rv.StatusCode()
		res.Write(internal.B(rv.Error()))
	default:
		res.statusCode = http.StatusInternalServerError
	}
}

func (ctx *RequestCtx) WriteStatus(status int) {
	ctx.WriteError(StatusError(status))
}
