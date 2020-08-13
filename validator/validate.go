package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
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

var ErrNotJsonRequest = &FormError{v: "not a json request"}
var jsonCt = []byte("application/json")

func Validate(ctx *fasthttp.RequestCtx, ptr interface{}) bool {
	_v := reflect.ValueOf(ptr).Elem()
	rules := GetRules(_v.Type())
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
					output.Error(ctx, newInvalidError(rule.form))
					return false
				}
				value = reflect.ValueOf(m)
			case _JsonArray:
				m, ok := rule.toJsonAry(val)
				if !ok {
					output.Error(ctx, newInvalidError(rule.form))
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
				output.Error(ctx, newNullError(rule.form))
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
			case _StringSlice:
				s, ok = rule.toStrSlice(ctx)
			case _JoinedIntSlice:
				s, ok = rule.toJoinedIntSlice(ctx)
			case _JoinedUintSlice:
				s, ok = rule.toJoinedUintSlice(ctx)
			case _JoinedBoolSlice:
				s, ok = rule.toJoinedBoolSlice(ctx)
			}

			if !ok {
				output.Error(ctx, newInvalidError(rule.form))
				return false
			}

			sV := reflect.ValueOf(s)
			if !sV.IsValid() {
				if rule.required {
					output.Error(ctx, newNullError(rule.form))
					return false
				} else {
					continue
				}
			}

			if !rule.checkSize(&sV) {
				output.Error(ctx, newNullError(rule.form))
				return false
			}
			return true
		}

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
			field.SetString(gotils.B2S(v))
		}
	}
	return true
}

func ValidateJson(ctx *fasthttp.RequestCtx, ptr interface{}) bool {
	if !bytes.Equal(ctx.Request.Header.ContentType(), jsonCt) {
		output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
		return false
	}

	err := json.Unmarshal(ctx.Request.Body(), ptr)
	if err != nil {
		output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
		return false
	}
	return true
}
