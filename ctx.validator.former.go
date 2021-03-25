package sha

import "github.com/zzztttkkk/sha/validator"

var _ validator.Former = _Former{}

type _Former struct{ *Request }

func (f _Former) HeaderValue(name string) ([]byte, bool) { return f.Request.header.Get(name) }

func (f _Former) URLParam(name string) ([]byte, bool) { return f.Request.URLParams.Get(name) }

func (f _Former) QueryValue(name string) ([]byte, bool) { return f.Request.QueryValue(name) }

func (f _Former) QueryValues(name string) [][]byte { return f.Request.QueryValues(name) }

func (f _Former) BodyValue(name string) ([]byte, bool) { return f.Request.BodyFormValue(name) }

func (f _Former) BodyValues(name string) [][]byte { return f.Request.BodyFormValues(name) }

func (f _Former) FormValue(name string) ([]byte, bool) {
	v, ok := f.Request.QueryValue(name)
	if ok {
		return v, true
	}
	return f.Request.BodyFormValue(name)
}

func (f _Former) FormValues(name string) [][]byte {
	v := f.Request.QueryValues(name)
	v = append(v, f.Request.BodyFormValues(name)...)
	return v
}

func (f _Former) HeaderValues(name string) [][]byte { return f.Request.Header().GetAll(name) }
