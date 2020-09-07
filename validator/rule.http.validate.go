package validator

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/output"
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

func _NewFormNullError(name string) *FormError {
	return &FormError{v: fmt.Sprintf("`%s` is required", name)}
}

func (r *_Rule) toFormError() *FormError {
	if len(r.info) > 0 {
		return &FormError{v: fmt.Sprintf("`%s` is invalid. %s", r.form, r.info)}
	}
	return &FormError{v: fmt.Sprintf("`%s` is invalid", r.form)}
}

var ErrNotJsonRequest = &FormError{v: "not a json request"}
var jsonCt = []byte("application/json")

//revive:disable:cyclomatic
func Validate(ctx *fasthttp.RequestCtx, ptr interface{}) bool {
	_v := reflect.ValueOf(ptr).Elem()
	rules := getRules(_v.Type())
	isJsonReq := bytes.HasPrefix(ctx.Request.Header.ContentType(), jsonCt)

	if rules.isJson {
		if isJsonReq {
			rule := rules.lst[0]
			field := _v.FieldByName(rule.field)
			val := ctx.Request.Body()
			var value reflect.Value
			switch rule.t {
			case _JsonObject:
				m, ok := rule.toJsonObj(val)
				if !ok {
					output.Error(ctx, rule.toFormError())
					return false
				}
				value = reflect.ValueOf(m)
			case _JsonArray:
				m, ok := rule.toJsonAry(val)
				if !ok {
					output.Error(ctx, rule.toFormError())
					return false
				}
				value = reflect.ValueOf(m)
			}
			field.Set(value)
			return true
		} else {
			output.Error(ctx, ErrNotJsonRequest)
			return false
		}
	}

	if isJsonReq {
		output.Error(ctx, ErrNotJsonRequest)
		return false
	}

	for _, rule := range rules.lst {
		val := ctx.FormValue(rule.form)
		if len(val) > 0 {
			val = bytes.TrimSpace(val)
		}

		if len(val) == 0 && len(rule.defaultV) > 0 {
			val = rule.defaultV
		}

		if len(val) == 0 {
			if rule.required {
				output.Error(ctx, _NewFormNullError(rule.form))
				return false
			}
			continue
		}

		field := _v.FieldByName(rule.field)

		if rule.isSlice {
			var ok bool
			var s interface{}

			switch rule.t {
			case _BoolSlice:
				s, ok = rule.toBoolSlice(ctx)
			case _IntSlice:
				s, ok = rule.toIntSlice(ctx)
			case _UintSlice:
				s, ok = rule.toUintSlice(ctx)
			case _FloatSlice:
				s, ok = rule.toFloatSlice(ctx)
			case _StringSlice:
				s, ok = rule.toStrSlice(ctx)
			}

			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}

			sV := reflect.ValueOf(s)
			if !sV.IsValid() {
				if rule.required {
					output.Error(ctx, _NewFormNullError(rule.form))
					return false
				} else {
					continue
				}
			}

			if !rule.checkSizeRange(&sV) {
				output.Error(ctx, _NewFormNullError(rule.form))
				return false
			}
			field.Set(sV)
			continue
		}

		switch rule.t {
		case _Bool:
			b, ok := rule.toBool(val)
			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}
			field.SetBool(b)
		case _Int64:
			v, ok := rule.toInt(val)
			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}
			field.SetInt(v)
		case _Uint64:
			v, ok := rule.toUint(val)
			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}
			field.SetUint(v)
		case _Float64:
			v, ok := rule.toFloat(val)
			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}
			field.SetFloat(v)
		case _Bytes:
			v, ok := rule.toBytes(val)
			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}
			field.SetBytes(v)
		case _String:
			v, ok := rule.toBytes(val)
			if !ok {
				output.Error(ctx, rule.toFormError())
				return false
			}
			field.SetString(gotils.B2S(v))
		}
	}
	return true
}

func ValidateJson(ctx *fasthttp.RequestCtx, ptr interface{}) bool {
	if !bytes.HasPrefix(ctx.Request.Header.ContentType(), jsonCt) {
		output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
		return false
	}

	err := jsonx.Unmarshal(ctx.Request.Body(), ptr)
	if err != nil {
		output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
		return false
	}
	return true
}
