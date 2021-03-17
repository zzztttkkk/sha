package sha

import (
	"bytes"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
)

func (ctx *RequestCtx) MustValidate(dist interface{}) {
	e := ctx.Validate(dist)
	if e != nil {
		panic(e)
	}
}

// revive:disable
// pointer -> interface
func (ctx *RequestCtx) Validate(dist interface{}) HTTPError {
	if err := validator.Validate(ctx, dist); err != nil {
		return err
	}
	return nil
}

//revive:enable

func (ctx *RequestCtx) ValidateJSON(dist interface{}) HTTPError {
	if !bytes.HasPrefix(ctx.Request.Header.ContentType(), utils.B(MIMEJson)) {
		return StatusError(StatusBadRequest)
	}
	if err := jsonx.Unmarshal(ctx.buf, dist); err != nil {
		return StatusError(StatusBadRequest)
	}
	return nil
}

func (ctx *RequestCtx) MustValidateJSON(dist interface{}) {
	err := ctx.ValidateJSON(dist)
	if err != nil {
		panic(err)
	}
}
