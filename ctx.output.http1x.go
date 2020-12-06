package suna

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"html/template"
	"io"
	"mime"
	"strconv"
)

var responseHeaderBufferPool = internal.NewBufferPoll(4096)

var newline = []byte("\r\n")
var headerKVSep = []byte(": ")

func (ctx *RequestCtx) KeepAlive() bool {
	req := &ctx.Request
	cv, _ := req.Header.Get(headerConnection)
	if string(inplaceLowercase(cv)) == "close" {
		return false
	}
	res := &ctx.Response
	cv, _ = res.Header.Get(headerConnection)
	if string(inplaceLowercase(cv)) == "close" {
		return false
	}
	return true
}

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

func (ctx *RequestCtx) WriteJSON(v interface{}) {
	ctx.Response.Header.Set(headerContentType, MIMEJson)

	encoder := json.NewEncoder(ctx)
	err := encoder.Encode(v)
	if err != nil {
		panic(err)
	}
}

func (ctx *RequestCtx) WriteHTML(v []byte) {
	ctx.Response.Header.Set(headerContentType, MIMEHtml)
	_, e := ctx.Write(v)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) WriteFile(f io.Reader, ext string) {
	ctx.Response.Header.SetStr("Content-Type", mime.TypeByExtension(ext))

	buf := make([]byte, 512, 512)
	for {
		l, e := f.Read(buf)
		if e != nil {
			panic(e)
		}
		_, e = ctx.Write(buf[:l])
		if e != nil {
			panic(e)
		}
	}
}

func (ctx *RequestCtx) WriteTemplate(t *template.Template, data interface{}) {
	ctx.Response.Header.Set(headerContentType, MIMEHtml)
	e := t.Execute(ctx, data)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) SetStatus(status int) {
	ctx.Response.statusCode = status
}

var ErrNotImpl = errors.New("suna: not implemented")

func (ctx *RequestCtx) Send() error {
	switch ctx.Request.version[0] {
	case '2':
		return ErrNotImpl
	}
	return ctx.sendHttp1xResponseBuffer()
}
