package unregister

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna.example/model"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/validator"

	"github.com/zzztttkkk/suna/output"
)

type Form struct {
	Secret string `validator:"L<64>"`
}

func HttpHandler(ctx *fasthttp.RequestCtx) {
	user := auth.MustGetUser(ctx)
	form := Form{}
	if !validator.Validate(ctx, &form) {
		return
	}

	Do(ctx, user.(*model.User), form.Secret)
	output.MsgOK(ctx, nil)
}
