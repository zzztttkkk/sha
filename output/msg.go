package output

import (
	"bytes"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/savsgio/gotils"
	"github.com/zzztttkkk/suna/jsonx"
	"log"

	"github.com/valyala/fasthttp"
)

type M map[string]interface{}

type Message struct {
	Errno  int         `json:"errno"`
	ErrMsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

var (
	strApplicationJSON       = []byte("application/json; charset=utf8")
	strApplicationJavascript = []byte("text/javascript; charset=utf8")
	emptyMsg                 = []byte(`{"errno":0,"errmsg":"","data":""}`)
	internalServerErrorMsg   = []byte(`{"errno":500,"errmsg":"internal server error","data":""}`)
	_JsonpLeft               = []byte(`(`)
	_JsonpRight              = []byte(`)`)
)

func Json(ctx *fasthttp.RequestCtx, data interface{}) {
	writer := _CtxCompressionWriter{raw: ctx}
	b, e := jsonx.Marshal(data)
	if e != nil {
		Error(ctx, e)
		return
	}

	var cb []byte
	if len(_JsonPCallbackParams) > 0 && gotils.B2S(ctx.Method()) == fasthttp.MethodGet {
		cb = ctx.FormValue(_JsonPCallbackParams)
	}
	if len(cb) < 1 {
		_, e = writer.Write(b)
		if e != nil {
			log.Printf("suna.output: %s\r\n", e.Error())
		}
		return
	}

	ctx.Response.Header.SetContentTypeBytes(strApplicationJavascript)
	buf := bytes.Buffer{}
	buf.Write(cb)
	buf.Write(_JsonpLeft)
	buf.Write(b)
	buf.Write(_JsonpRight)

	_, e = writer.Write(buf.Bytes())
	if e != nil {
		log.Printf("suna.output: %s\r\n", e.Error())
	}
}

func Msg(ctx *fasthttp.RequestCtx, code int, data interface{}) {
	ctx.SetContentTypeBytes(strApplicationJSON)
	ctx.SetStatusCode(code)

	if data == nil {
		_, err := ctx.Write(emptyMsg)
		if err != nil {
			log.Printf("suna.output: ctx write error, %s\r\n", err.Error())
		}
		return
	}

	msg := Message{Data: data}
	Json(ctx, &msg)
}

func MsgOK(ctx *fasthttp.RequestCtx, data interface{}) {
	Msg(ctx, fasthttp.StatusOK, data)
}

func ErrorAndErrorStack(ctx *fasthttp.RequestCtx, err error) string {
	ctx.Response.ResetBody()
	ctx.SetContentTypeBytes(strApplicationJSON)

	var code int

	switch v := err.(type) {
	case Err:
		code = v.StatusCode()
		ctx.SetStatusCode(v.StatusCode())
		Json(ctx, v.Message())
	default:
		code = fasthttp.StatusInternalServerError
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		_, _ = ctx.Write(internalServerErrorMsg)
	}

	if code > 499 {
		e := errors.Wrap(err, 1)
		return fmt.Sprintf("%s\n%s\n", e.Error(), e.ErrorStack())
	}
	return ""
}

func Error(ctx *fasthttp.RequestCtx, err error) {
	stack := ErrorAndErrorStack(ctx, err)
	if len(stack) > 1 {
		log.Print(stack)
	}
}
