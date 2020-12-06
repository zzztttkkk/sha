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
	valMap  map[interface{}]ErrorHandler
}

func (r *_Recover) doRecover(ctx *RequestCtx) {
	v := recover()
	if v == nil {
		return
	}

	ctx.Response.ResetBodyBuffer()

	fn := r.valMap[v]
	if fn != nil {
		fn(ctx, v)
		return
	}

	if len(r.typeMap) > 0 {
		t := reflect.TypeOf(v)
		fn = r.typeMap[t]
		if fn != nil {
			fn(ctx, v)
			return
		}
	}

	logStack := true
	switch rv := v.(type) {
	case HttpResponseError:
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
	case HttpError:
		if rv.StatusCode() < 500 {
			logStack = false
		}
		ctx.SetStatus(rv.StatusCode())
		_, _ = ctx.WriteString(rv.Error())
	default:
		ctx.SetStatus(http.StatusInternalServerError)
	}

	if logStack {
		log.Print(internal.Stacks(v, 2, 20))
	}
}
