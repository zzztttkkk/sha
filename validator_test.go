package sha

import (
	"github.com/zzztttkkk/sha/auth"
	"testing"
)

type TestForm struct {
	Name string  `validator:",L=3-20"`
	Nums []int64 `validator:",V=0-9,S=3"`
}

func (TestForm) Default(fieldName string) interface{} {
	switch fieldName {
	case "Nums":
		return []int64{1, 2, 3}
	}
	return nil
}

func TestRequestCtx_Validate(t *testing.T) {
	mux := NewMux("", nil)
	server := Default(mux)

	mux.HTTPWithForm(
		"get",
		"/",
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				form := TestForm{}
				panic(auth.ErrUnauthenticatedOperation)
				ctx.MustValidate(&form)
				_, _ = ctx.WriteString("OK")
			},
		),
		TestForm{},
	)

	mux.HandleDoc("get", "/doc")

	server.ListenAndServe()
}
