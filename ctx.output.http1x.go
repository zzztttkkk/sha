package sha

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/internal"
	"html/template"
	"io"
	"mime"
	"strconv"
)

var newline = []byte("\r\n")
var headerKVSep = []byte(": ")

func (ctx *RequestCtx) KeepAlive() bool {
	const closeStr = "close"

	req := &ctx.Request
	hc := internal.B(HeaderConnection)
	cv, _ := req.Header.Get(hc)
	if string(inplaceLowercase(cv)) == closeStr {
		return false
	}
	res := &ctx.Response
	cv, _ = res.Header.Get(hc)
	if string(inplaceLowercase(cv)) == closeStr {
		return false
	}
	return true
}

func (ctx *RequestCtx) sendHttp1xResponseBuffer() error {
	res := &ctx.Response
	if res.compressWriter != nil {
		_ = res.compressWriter.Flush()
	}

	res.Header.SetContentLength(int64(len(res.buf.Data)))
	if !res.headerWritten {
		err := ctx.writeHttp1xHeader()
		if err != nil {
			return err
		}
	}
	if _, err := ctx.conn.Write(newline); err != nil {
		return err
	}

	_, err := ctx.conn.Write(ctx.Response.buf.Data)
	return err
}

var ErrNilContentLength = fmt.Errorf("sha: nil content length")
var ErrUnknownResponseStatusCode = fmt.Errorf("sha: unknown response status code")
var responseHeaderBufferPool = internal.NewBufferPoll(4096)

func (ctx *RequestCtx) writeHttp1xHeader() error {
	res := &ctx.Response
	res.headerWritten = true

	_, exists := res.Header.Get(internal.B(HeaderContentLength))
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
			encodeHeaderValue(v, &buf.Data)
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
	ctx.Response.Header.SetContentType(MIMEJson)

	encoder := json.NewEncoder(ctx)
	err := encoder.Encode(v)
	if err != nil {
		panic(err)
	}
}

func (ctx *RequestCtx) WriteHTML(v []byte) {
	ctx.Response.Header.SetContentType(MIMEHtml)
	_, e := ctx.Write(v)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) WriteFile(f io.Reader, ext string) {
	ctx.Response.Header.SetContentType(internal.B(mime.TypeByExtension(ext)))

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
	ctx.Response.Header.SetContentType(MIMEHtml)
	e := t.Execute(ctx, data)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) SetStatus(status int) {
	ctx.Response.statusCode = status
}

var ErrNotImpl = errors.New("sha: not implemented")

func (ctx *RequestCtx) Send() error {
	switch ctx.Request.version[5] {
	case '2':
		return ErrNotImpl
	}
	return ctx.sendHttp1xResponseBuffer()
}
