package sha

import (
	"github.com/zzztttkkk/sha/internal"
	"log"
	"net/http"
	"reflect"
)

type ErrorHandler func(ctx *RequestCtx, v interface{})

type _Recover struct {
	typeMap map[reflect.Type]ErrorHandler
	valMap  map[error]ErrorHandler
}

func (r *_Recover) RecoverByType(t reflect.Type, fn ErrorHandler) {
	if r.typeMap == nil {
		r.typeMap = map[reflect.Type]ErrorHandler{}
	}
	r.typeMap[t] = fn
}

func (r *_Recover) RecoverByErr(v error, fn ErrorHandler) {
	if r.valMap == nil {
		r.valMap = map[error]ErrorHandler{}
	}
	r.valMap[v] = fn
}

var (
	errType        = reflect.TypeOf((*error)(nil)).Elem()
	httpResErrType = reflect.TypeOf((*HttpResponseError)(nil)).Elem()
	httpErrType    = reflect.TypeOf((*HttpError)(nil)).Elem()
)

func (r *_Recover) doRecover(ctx *RequestCtx) {
	v := recover()
	if v == nil {
		return
	}

	ctx.Response.ResetBodyBuffer()

	vt := reflect.TypeOf(v)

	if vt.ConvertibleTo(errType) {
		fn := r.valMap[v.(error)]
		if fn != nil {
			fn(ctx, v)
			return
		}
	}

	if len(r.typeMap) > 0 {
		fn := r.typeMap[vt]
		if fn != nil {
			fn(ctx, v)
			return
		}
	}

	logStack := true

	if vt.ConvertibleTo(httpResErrType) {
		rv := v.(HttpResponseError)
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.SetStatus(rv.StatusCode())
		rv.Header(&ctx.Response.Header)
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
