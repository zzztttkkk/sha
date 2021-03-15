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
						return func(ctx *RequestCtx, _ interface{}) { ctx.SetStatus(sc) }
					}(v),
				)
			}
		},
	)
}

var (
	errType        = reflect.TypeOf((*error)(nil)).Elem()
	httpResErrType = reflect.TypeOf((*HttpResponseError)(nil)).Elem()
	httpErrType    = reflect.TypeOf((*HttpError)(nil)).Elem()
)

func doRecover(ctx *RequestCtx, v interface{}) {
	ctx.Response.ResetBodyBuffer()

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
		rv := v.(HttpResponseError)
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.SetStatus(rv.StatusCode())
		rv.WriteHeader(&ctx.Response.Header)
		_, _ = ctx.Write(rv.Body())
	} else if vt.ConvertibleTo(httpErrType) {
		rv := v.(HttpError)
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.SetStatus(rv.StatusCode())
		_, _ = ctx.WriteString(rv.Error())
	} else {
		ctx.SetStatus(http.StatusInternalServerError)
	}

	if logStack {
		log.Print(internal.Stacks(v, 2, 20))
	}
}
