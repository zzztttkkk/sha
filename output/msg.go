package output

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/valyala/fasthttp"
	"log"
)

type M map[string]interface{}

type Message struct {
	Errno  int         `json:"errno"`
	ErrMsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

var (
	strApplicationJSON     = []byte("application/json; charset=utf8")
	emptyMsg               = []byte(`{"errno":0,"errmsg":"","data":""}`)
	internalServerErrorMsg = []byte(`{"errno":500,"errmsg:"internal server error","data":""}`)
)

func toJson(ctx *fasthttp.RequestCtx, data interface{}) {
	encoder := json.NewEncoder(ctx)
	err := encoder.Encode(data)
	if err == nil {
		return
	}
	log.Printf("snow.output: json encode error, %s\r\n", err.Error())
}

func Msg(ctx *fasthttp.RequestCtx, code int, data interface{}) {
	ctx.SetContentTypeBytes(strApplicationJSON)
	ctx.SetStatusCode(code)

	if data == nil {
		_, err := ctx.Write(emptyMsg)
		if err != nil {

		}
		return
	}

	msg := Message{Data: data}
	toJson(ctx, &msg)
}

func MsgOk(ctx *fasthttp.RequestCtx, data interface{}) {
	Msg(ctx, fasthttp.StatusOK, data)
}

func Error(ctx *fasthttp.RequestCtx, err error) {
	ctx.Response.ResetBody()
	ctx.SetContentTypeBytes(strApplicationJSON)

	log.Println(errors.Wrap(err, 1).ErrorStack())

	switch v := err.(type) {
	case HttpError:
		ctx.SetStatusCode(v.StatusCode())
		toJson(ctx, v.Message())
	default:
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		_, _ = ctx.Write(internalServerErrorMsg)
	}
}

func StdError(ctx *fasthttp.RequestCtx, code int) {
	err := StdErrors[code]
	if err == nil {
		err = StdErrors[fasthttp.StatusInternalServerError]
	}
	Error(ctx, err)
}
