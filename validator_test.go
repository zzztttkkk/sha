package suna

import (
	"testing"
)

func TestRequestCtx_Validate(t *testing.T) {
	s := Default()

	mux := NewMux("", nil)

	type Form struct {
		Name string  `validator:"L<3-20>"`
		Nums []int64 `validator:"V<0-9>;S<5>"`
	}

	mux.AddHandlerWithForm(
		"get",
		"/",
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				form := Form{}
				ctx.MustValidate(&form)
				_, _ = ctx.WriteString("OK")
			},
		),
		Form{},
	)

	mux.HandleDoc("get", "/doc")

	s.Handler = mux

	s.ListenAndServe()
}
