package login

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/session"
	"github.com/zzztttkkk/suna/validator"
	"time"
)

type Form struct {
	Name     []byte `validator:"F<username>;L<3-20>"`
	Password []byte `validator:"R<password>"`
	Keep     int64  `validator:"D<0>;V<0-30>;optional"`
}

func HttpHandler(ctx *fasthttp.RequestCtx) {
	s := session.New(ctx)
	if !s.CaptchaVerify(ctx) {
		output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
		return
	}

	form := Form{}
	if !validator.Validate(ctx, &form) {
		return
	}

	var maxage = time.Hour * time.Duration(24*form.Keep)
	if form.Keep < 1 {
		maxage = time.Hour * 3
	}

	token, err := Do(ctx, form.Name, form.Password, int64(maxage/time.Second))
	if err != nil {
		output.Error(ctx, err)
		return
	}
	output.MsgOK(ctx, token)
}
