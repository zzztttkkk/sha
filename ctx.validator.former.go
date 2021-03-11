package sha

import "github.com/zzztttkkk/sha/validator"

var _ validator.Former = (*RequestCtx)(nil)

func (ctx *RequestCtx) URLParam(name string) ([]byte, bool) { return ctx.Request.URLParams.Get(name) }

func (ctx *RequestCtx) QueryValue(name string) ([]byte, bool) { return ctx.Request.QueryValue(name) }

func (ctx *RequestCtx) QueryValues(name string) [][]byte { return ctx.Request.QueryValues(name) }

func (ctx *RequestCtx) BodyValue(name string) ([]byte, bool) { return ctx.Request.BodyFormValue(name) }

func (ctx *RequestCtx) BodyValues(name string) [][]byte { return ctx.Request.BodyFormValues(name) }

func (ctx *RequestCtx) FormValue(name string) ([]byte, bool) {
	v, ok := ctx.Request.QueryValue(name)
	if ok {
		return v, true
	}
	return ctx.Request.BodyFormValue(name)
}

func (ctx *RequestCtx) FormValues(name string) [][]byte {
	v := ctx.Request.QueryValues(name)
	v = append(v, ctx.Request.BodyFormValues(name)...)
	return v
}

func (ctx *RequestCtx) HeaderValue(key string) ([]byte, bool) { return ctx.Request.Header.Get(key) }

func (ctx *RequestCtx) HeaderValues(key string) [][]byte { return ctx.Request.Header.GetAll(key) }

func (ctx *RequestCtx) CookieValue(key string) ([]byte, bool) { return ctx.Request.CookieValue(key) }

// others
func (ctx *RequestCtx) File(name []byte) *FormFile { return ctx.Request.Files().Get(name) }

func (ctx *RequestCtx) Files(name []byte) []*FormFile { return ctx.Request.Files().GetAll(name) }

func (ctx *RequestCtx) BodyRaw() []byte { return ctx.buf }
