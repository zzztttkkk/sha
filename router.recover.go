package sha

import (
	"github.com/zzztttkkk/sha/internal"
	"log"
	"net/http"
	"reflect"
)

type ErrorHandler func(ctx *RequestCtx, v interface{})

var errTypeMap = map[reflect.Type]ErrorHandler{}
var errValMap = map[error]ErrorHandler{}

func RecoverByType(t reflect.Type, fn ErrorHandler) { errTypeMap[t] = fn }

func RecoverByErr(v error, fn ErrorHandler) { errValMap[v] = fn }

func init() {
	serverPrepareFunc = append(
		serverPrepareFunc,
		func(server *Server) {
			_, ok := server.Handler.(*Mux)
			if !ok {
				return
			}
			for k, v := range internal.ErrorStatusByValue {
				RecoverByErr(
					k,
					func(sc int) ErrorHandler {
						return func(ctx *RequestCtx, _ interface{}) { ctx.Response.SetStatusCode(sc) }
					}(v),
				)
			}
		},
	)
}

var (
	errType                        = reflect.TypeOf((*error)(nil)).Elem()
	httpResErrType                 = reflect.TypeOf((*HTTPResponseError)(nil)).Elem()
	httpErrType                    = reflect.TypeOf((*HTTPError)(nil)).Elem()
	CallersFramesSkip          int = 0
	CallersFramesSize          int = 20
	InternalServerErrorMessage     = `<h1>Oops, an unknown internal server error occurred.`
)

func defaultRecover(ctx *RequestCtx, v interface{}) {
	ctx.Response.ResetBody()
	ctx.Response.Header().Del(HeaderContentType)
	ctx.Response.Header().Del(HeaderContentLength)

	vt := reflect.TypeOf(v)

	if vt.ConvertibleTo(errType) {
		fn := errValMap[v.(error)]
		if fn != nil {
			fn(ctx, v)
			return
		}
	}

	if fn := errTypeMap[vt]; fn != nil {
		fn(ctx, v)
		return
	}

	logStack := true

	if vt.ConvertibleTo(httpResErrType) {
		rv := v.(HTTPResponseError)
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.Response.SetStatusCode(rv.StatusCode())
		rv.WriteHeader(ctx.Response.Header())
		_, _ = ctx.Write(rv.Body())
	} else if vt.ConvertibleTo(httpErrType) {
		rv := v.(HTTPError)
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.Response.SetStatusCode(rv.StatusCode())
		_ = ctx.WriteString(rv.Error())
	} else {
		ctx.Response.SetStatusCode(http.StatusInternalServerError)
		ctx.Response.Header().SetContentType(MIMEHtml)
		_ = ctx.WriteString(InternalServerErrorMessage)
	}

	if logStack {
		log.Print(internal.LoadCallersFrames(v, CallersFramesSkip, CallersFramesSize))
	}
}
