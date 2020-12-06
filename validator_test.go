package suna

import (
	"testing"
)

type TestForm struct {
	Name string  `validator:",L=3-20"`
	Nums []int64 `validator:",V=0-9,S=3"`
}

func (TestForm) Default(fn string) interface{} {
	switch fn {
	case "Nums":
		return []int64{1, 2, 3}
	}
	return nil
}

func TestRequestCtx_Validate(t *testing.T) {
	s := Default(nil)

	mux := NewMux("", nil)

	mux.RESTWithForm(
		"get",
		"/",
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				form := TestForm{}
				ctx.MustValidate(&form)
				_, _ = ctx.WriteString("OK")
			},
		),
		TestForm{},
	)

	mux.HandleDoc("get", "/doc")

	s.Handler = mux

	s.ListenAndServe()
}
