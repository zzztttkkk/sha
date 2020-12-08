package suna

import (
	"github.com/zzztttkkk/suna/internal"
	"log"
	"net/http"
	"reflect"
)

type ErrorHandler func(ctx *RequestCtx, v interface{})

type _Recover struct {
	typeMap map[reflect.Type]ErrorHandler
	valMap  map[error]ErrorHandler
}

var (
	errType     = reflect.TypeOf((*error)(nil)).Elem()
	hResErrType = reflect.TypeOf((*HttpResponseError)(nil)).Elem()
	hErrType    = reflect.TypeOf((*HttpError)(nil)).Elem()
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

	if vt.ConvertibleTo(hResErrType) {
		rv := v.(HttpResponseError)
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.SetStatus(rv.StatusCode())
		header := rv.Header()
		if header != nil {
			header.EachItem(
				func(k, v []byte) bool {
					ctx.Response.Header.Set(k, v)
					return true
				},
			)
		}
		_, _ = ctx.Write(rv.Body())
	} else if vt.ConvertibleTo(hErrType) {
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
