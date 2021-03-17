package sha

import (
	"github.com/zzztttkkk/sha/validator"
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
	mux := NewMux(nil)
	server := Default(mux)

	mux.HTTPWithOptions(
		&HandlerOptions{
			Document: validator.NewDocument(TestForm{}, validator.Undefined),
		},
		"get",
		"/",
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				form := TestForm{}
				ctx.MustValidate(&form)
				_, _ = ctx.WriteString("OK")
			},
		),
	)

	server.ListenAndServe()
}
