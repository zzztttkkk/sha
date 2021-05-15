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

// ValidateForm error pointer ->  error interface
func (ctx *RequestCtx) ValidateForm(dist interface{}) HTTPError {
	if err := validator.BindAndValidateForm(_Former{&ctx.Request}, dist); err != nil {
		return err
	}
	return nil
}

func (ctx *RequestCtx) ValidateJSON(dist interface{}) HTTPError {
	if !bytes.HasPrefix(ctx.Request.Header().ContentType(), utils.B(MIMEJson)) {
		return StatusError(StatusUnsupportedMediaType)
	}

	body := ctx.Request._HTTPPocket.body
	if body == nil {
		return StatusError(StatusBadRequest)
	}
	if err := jsonx.Unmarshal(body.Bytes(), dist); err != nil {
		return StatusError(StatusBadRequest)
	}
	if err := validator.ValidateStruct(dist); err != nil {
		return err
	}
	return nil
}

func (ctx *RequestCtx) MustValidateJSON(dist interface{}) {
	err := ctx.ValidateJSON(dist)
	if err != nil {
		panic(err)
	}
}

func (ctx *RequestCtx) Validate(dist interface{}) HTTPError {
	err := ctx.ValidateJSON(dist)
	if err != nil {
		if err.StatusCode() == StatusUnsupportedMediaType {
			return ctx.ValidateForm(dist)
		}
		return err
	}
	return nil
}

func (ctx *RequestCtx) MustValidate(dist interface{}) {
	if err := ctx.Validate(dist); err != nil {
		panic(err)
	}
}
