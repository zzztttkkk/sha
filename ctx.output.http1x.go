package sha

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"html/template"
	"io"
	"mime"
	"strconv"
)

func (ctx *RequestCtx) sendHttp1xResponseBuffer() error {
	res := &ctx.Response
	if res.compressWriter != nil {
		_ = res.compressWriter.Flush()
	}

	size := int64(len(res.bodyBuf.Data))

	res.Header.SetContentLength(size)
	err := ctx.writeHttp1xHeader()
	if err != nil {
		return err
	}

	_, err = res.sendBuf.Write(ctx.Response.bodyBuf.Data)
	if err != nil {
		return err
	}
	return res.sendBuf.Flush()
}

const (
	EndLine     = "\r\n"
	headerKVSep = ": "
)

var ErrUnknownResponseStatusCode = fmt.Errorf("sha: unknown response status code")

func (ctx *RequestCtx) writeHttp1xHeader() error {
	res := &ctx.Response

	if res.statusCode < 1 {
		res.statusCode = 200
	}

	statusTxt := statusTextMap[res.statusCode]
	if len(statusTxt) < 1 {
		return ErrUnknownResponseStatusCode
	}

	res.headerBuf = append(res.headerBuf, ctx.protocol.Version...)
	res.headerBuf = append(res.headerBuf, ' ')
	res.headerBuf = append(res.headerBuf, strconv.FormatInt(int64(res.statusCode), 10)...)
	res.headerBuf = append(res.headerBuf, ' ')
	res.headerBuf = append(res.headerBuf, statusTxt...)
	res.headerBuf = append(res.headerBuf, EndLine...)

	res.Header.EachItem(
		func(item *utils.KvItem) bool {
			res.headerBuf = append(res.headerBuf, item.Key...)
			res.headerBuf = append(res.headerBuf, headerKVSep...)
			encodeHeaderValue(item.Val, &res.headerBuf)
			res.headerBuf = append(res.headerBuf, EndLine...)
			return true
		},
	)

	res.headerBuf = append(res.headerBuf, EndLine...)
	_, e := res.sendBuf.Write(res.headerBuf)
	return e
}

func (ctx *RequestCtx) Write(p []byte) (int, error) {
	return ctx.Response.Write(p)
}

func (ctx *RequestCtx) WriteString(s string) (int, error) {
	return ctx.Write(utils.B(s))
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
	ctx.Response.Header.SetContentType(mime.TypeByExtension(ext))

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
