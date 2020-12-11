package sha

import (
	"bytes"
	"encoding/json"
	"github.com/zzztttkkk/sha/validator"
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

func (ctx *RequestCtx) ValidateJSON(dist interface{}) HttpError {
	if !bytes.HasPrefix(ctx.Request.Header.ContentType(), MIMEJson) {
		return StatusError(StatusBadRequest)
	}
	if err := json.Unmarshal(ctx.buf, dist); err != nil {
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
