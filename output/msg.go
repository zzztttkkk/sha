package output

import (
	"bytes"
	"github.com/zzztttkkk/suna/jsonx"
	"log"

	"github.com/go-errors/errors"
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
	_MethodGet               = []byte("GET")
)

func Json(ctx *fasthttp.RequestCtx, data interface{}) {
	var cb []byte
	if len(_JsonPCallbackParams) > 0 && bytes.Equal(ctx.Method(), _MethodGet) {
		cb = ctx.FormValue(_JsonPCallbackParams)
	}
	if len(cb) < 1 {
		_ = jsonx.EncodeTo(data, ctx, nil)
		return
	}

	ctx.Response.Header.SetContentTypeBytes(strApplicationJavascript)
	_, _ = ctx.Write(cb)
	_, _ = ctx.Write(_JsonpLeft)
	_ = jsonx.EncodeTo(data, ctx, nil)
	_, _ = ctx.Write(_JsonpRight)
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

func Error(ctx *fasthttp.RequestCtx, v interface{}) {
	ctx.Response.ResetBody()
	ctx.SetContentTypeBytes(strApplicationJSON)
	switch rv := v.(type) {
	case MessageErr:
		ctx.SetStatusCode(rv.StatusCode())
		Json(ctx, rv.Message())
	default:
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		_, _ = ctx.Write(internalServerErrorMsg)
	}
}

func ErrorStack(e interface{}, skip int) string {
	return errors.Wrap(e, skip).ErrorStack()
}
