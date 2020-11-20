package suna

import (
	"github.com/zzztttkkk/suna/validator"
)

func (ctx *RequestCtx) MustValidate(dist interface{}) {
	e := ctx.Validate(dist)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) Validate(dist interface{}) HttpError {
	e, isNil := validator.Validate(ctx, dist)
	if isNil {
		return nil
	}
	return e
}

type _ValidateHandler struct {
	rules validator.Rules
	fn    func(ctx *RequestCtx)
}

var _ DocedRequestHandler = &_ValidateHandler{}

func (h *_ValidateHandler) Document() string {
	return h.rules.String()
}

func (h *_ValidateHandler) Handle(ctx *RequestCtx) {
	h.fn(ctx)
}
