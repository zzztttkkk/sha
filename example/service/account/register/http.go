package register

import (
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/session"
	"github.com/zzztttkkk/suna/validator"
)

type Form struct {
	Name     []byte `validator:"F<username>;L<3-20>"`
	Password []byte `validator:"R<password>"`
}

func HttpHandler(ctx *fasthttp.RequestCtx) {
	s := session.New(ctx)
	if !s.CaptchaVerify(ctx) {
		return
	}

	form := Form{}
	if !validator.Validate(ctx, &form) {
		return
	}

	uid, skey := Do(ctx, form.Name, form.Password)
	output.MsgOK(ctx, output.M{"uid": uid, "secret": gotils.B2S(skey)})
}
