package sha

import (
	"github.com/zzztttkkk/sha/validator"
	"testing"
)

type TestForm struct {
	Name string  `validator:",L=3-20"`
	Nums []int64 `validator:",V=0-9,S=3"`
}

func (TestForm) Default(fieldName string) func() interface{} {
	switch fieldName {
	case "Nums":
		return func() interface{} { return []int64{1, 2, 3, 678} }
	}
	return nil
}

func TestRequestCtx_Validate(t *testing.T) {
	mux := NewMux(nil)
	mux.HTTPWithOptions(
		&RouteOptions{Document: validator.NewDocument(TestForm{}, nil)},
		"get",
		"/",
		RequestCtxHandlerFunc(
			func(ctx *RequestCtx) {
				form := TestForm{}
				ctx.MustValidateForm(&form)
				_ = ctx.WriteString("OK")
			},
		),
	)
	ListenAndServe("", mux)
}
