package suna

import (
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/validator"
	"net/http"
	"reflect"
)

type FormError struct {
	ErrMessage string `json:"errmsg"`
	Field      string `json:"field"`
}

func (fe *FormError) Error() string {
	return fmt.Sprintf("FormError: field `%s`", fe.Field)
}

func (fe *FormError) StatusCode() int {
	return http.StatusBadRequest
}

func (fe *FormError) Message() []byte {
	return JsonMustMarshal(fe)
}

func (fe *FormError) Header() *Header {
	return nil
}

var _ HttpError = &FormError{}

func (ctx *RequestCtx) MustValidate(dist interface{}) {
	e := ctx.Validate(dist)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) Validate(dist interface{}) HttpError {
	v := reflect.ValueOf(dist).Elem()

	var field reflect.Value
	var ok bool

	for _, rule := range *validator.GetRules(v.Type()) {
		field = v
		for _, index := range rule.FieldIndex {
			field = field.Field(index)
		}
		if rule.IsSlice {
			ok = ctx.validateSlice(rule, &field)
		} else {
			ok = ctx.validateOne(rule, &field)
		}
		if !ok {
			err := &FormError{Field: internal.S(rule.FormName)}
			if len(rule.ErrMessage) > 0 {
				err.ErrMessage = rule.ErrMessage
			}
			return err
		}
	}
	return nil
}

func (ctx *RequestCtx) validateOne(rule *validator.Rule, field *reflect.Value) bool {
	req := &ctx.Request

	var fv []byte
	var ok bool
	var fieldValue interface{}

	if len(rule.PathParamsName) > 0 {
		fv, ok = req.Params.Get(rule.PathParamsName)
	} else {
		fv, ok = ctx.FormValue(rule.FormName)
	}

	if !ok {
		if rule.Default != nil {
			fv = *rule.Default
		} else {
			if rule.IsRequired {
				return false
			}
			return true
		}
	}

	switch rule.Type {
	case validator.Bool:
		fieldValue, ok = rule.ToBool(fv)
	case validator.Int64:
		fieldValue, ok = rule.ToInt(fv)
	case validator.Uint64:
		fieldValue, ok = rule.ToUint(fv)
	case validator.Float64:
		fieldValue, ok = rule.ToFloat(fv)
	case validator.Bytes:
		fieldValue, ok = rule.ToBytes(fv)
	case validator.String:
		fieldValue, ok = rule.ToString(fv)
	default:
		panic(StdHttpErrors[http.StatusInternalServerError])
	}
	if !ok {
		return false
	}
	field.Set(reflect.ValueOf(fieldValue))
	return true
}

func (ctx *RequestCtx) validateSlice(rule *validator.Rule, field *reflect.Value) bool {
	var fieldValue interface{}

	switch rule.Type {
	case validator.BoolSlice:
		var lst []bool
		for _, bs := range ctx.FormValues(rule.FormName) {
			a, b := rule.ToBool(bs)
			if !b {
				return false
			}
			lst = append(lst, a)
		}
		fieldValue = lst
	case validator.IntSlice:
		var lst []int64
		for _, bs := range ctx.FormValues(rule.FormName) {
			a, b := rule.ToInt(bs)
			if !b {
				return false
			}
			lst = append(lst, a)
		}
		fieldValue = lst
	case validator.UintSlice:
		var lst []uint64
		for _, bs := range ctx.FormValues(rule.FormName) {
			a, b := rule.ToUint(bs)
			if !b {
				return false
			}
			lst = append(lst, a)
		}
		fieldValue = lst
	case validator.FloatSlice:
		var lst []float64
		for _, bs := range ctx.FormValues(rule.FormName) {
			a, b := rule.ToFloat(bs)
			if !b {
				return false
			}
			lst = append(lst, a)
		}
		fieldValue = lst
	case validator.StringSlice:
		var lst []string
		for _, bs := range ctx.FormValues(rule.FormName) {
			a, b := rule.ToString(bs)
			if !b {
				return false
			}
			lst = append(lst, a)
		}
		fieldValue = lst
	case validator.BytesSlice:
		var lst [][]byte
		for _, bs := range ctx.FormValues(rule.FormName) {
			a, b := rule.ToBytes(bs)
			if !b {
				return false
			}
			lst = append(lst, a)
		}
		fieldValue = lst
	default:
		panic(StdHttpErrors[http.StatusInternalServerError])
	}

	v := reflect.ValueOf(fieldValue)
	if !rule.ValidateSliceValue(&v) {
		return false
	}
	field.Set(v)
	return true
}

type _ValidateHandler struct {
	rules validator.Rules
	fn    func(ctx *RequestCtx)
}

var _ DocedRequestHandler = &_ValidateHandler{}

func (h *_ValidateHandler) Document() string {
	return h.rules.String()
}

func (h *_ValidateHandler) Handle(ctx *RequestCtx) {
	h.fn(ctx)
}
