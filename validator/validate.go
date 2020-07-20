package validator

import (
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

type FormError struct {
	v string
}

func (e *FormError) Error() string {
	return e.v
}

func (e *FormError) StatusCode() int {
	return fasthttp.StatusBadRequest
}

func (e *FormError) Message() *output.Message {
	return &output.Message{
		Errno:  -1,
		ErrMsg: e.v,
		Data:   nil,
	}
}

func newNullError(name string) *FormError {
	return &FormError{v: fmt.Sprintf("`%s` is required", name)}
}

func newInvalidError(name string) *FormError {
	return &FormError{v: fmt.Sprintf("`%s` is invalid", name)}
}

func Validate(ctx *fasthttp.RequestCtx, ptr interface{}) bool {
	_v := reflect.ValueOf(ptr).Elem()
	for _, rule := range getRules(_v.Type()) {
		val := ctx.FormValue(rule.form)
		if val != nil {
			val = bytes.TrimSpace(val)
		}

		if len(val) == 0 && len(rule.defaultV) > 0 {
			val = rule.defaultV
		}

		if len(val) == 0 {
			if rule.required {
				output.Error(ctx, newNullError(rule.form))
				return false
			}
			continue
		}

		field := _v.FieldByName(rule.field)
		switch rule.t {
		case _Bool:
			b, ok := rule.toBool(val)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.SetBool(b)
		case _Int64:
			v, ok := rule.toI64(val)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.SetInt(v)
		case _Uint64:
			v, ok := rule.toUI64(val)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.SetUint(v)
		case _Bytes:
			v, ok := rule.toBytes(val)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.SetBytes(v)
		case _String:
			v, ok := rule.toBytes(val)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.SetString(utils.B2s(v))
		case _JsonObject:
			m, ok := rule.toJsonObj(val)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.Set(reflect.ValueOf(m))
		case _BoolSlice:
			s, ok := rule.toBoolSlice(ctx)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.Set(reflect.ValueOf(s))
		case _IntSlice:
			s, ok := rule.toIntSlice(ctx)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.Set(reflect.ValueOf(s))
		case _UintSlice:
			s, ok := rule.toUintSlice(ctx)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.Set(reflect.ValueOf(s))
		case _StringSlice:
			s, ok := rule.toStrSlice(ctx)
			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}
			field.Set(reflect.ValueOf(s))
		}
	}

	return true
}
