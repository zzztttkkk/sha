package login

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/middleware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/validator"
)

type Form struct {
	Name     []byte `validator:":F<username>;L<3-20>"`
	Password []byte `validator:":R<password>"`
	Days     int    `validator:":optional;D<0>;V<0-30>"`
}

func Handler(ctx *fasthttp.RequestCtx) {
	session := ctxs.Session(ctx)
	if !session.CaptchaVerify(ctx) {
		output.StdError(ctx, fasthttp.StatusBadRequest)
		return
	}

	form := Form{}
	if !validator.Validate(ctx, &form) {
		return
	}
	token, err := DoLogin(ctx, form.Name, form.Password, form.Days)
	if err != nil {
		output.Error(ctx, err)
		return
	}
	output.MsgOk(ctx, token)
}
