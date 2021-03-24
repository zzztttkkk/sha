package sha

import (
	"bytes"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
)

func (ctx *RequestCtx) MustValidateForm(dist interface{}) {
	e := ctx.ValidateForm(dist)
	if e != nil {
		panic(e)
	}
}

// revive:disable

// pointer -> interface
func (ctx *RequestCtx) ValidateForm(dist interface{}) HTTPError {
	if err := validator.BindAndValidateForm(_Former{&ctx.Request}, dist); err != nil {
		return err
	}
	return nil
}

func (ctx *RequestCtx) ValidateJSON(dist interface{}) HTTPError {
	if !bytes.HasPrefix(ctx.Request.Header.ContentType(), utils.B(MIMEJson)) {
		return StatusError(StatusBadRequest)
	}
	if err := jsonx.Unmarshal(ctx.buf, dist); err != nil {
		return StatusError(StatusBadRequest)
	}
	if err := validator.ValidateStruct(dist); err != nil {
		return err
	}
	return nil
}

//revive:enable

func (ctx *RequestCtx) MustValidateJSON(dist interface{}) {
	err := ctx.ValidateJSON(dist)
	if err != nil {
		panic(err)
	}
}
