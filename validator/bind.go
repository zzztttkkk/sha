package validator

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"net/http"
	"reflect"
)

type Former interface {
	PathParam(name []byte) ([]byte, bool)
	FormValue(name []byte) ([]byte, bool)
	FormValues(name []byte) [][]byte
}

type _FormErrorType int

func (v _FormErrorType) String() string {
	switch v {
	case MissingRequired:
		return "missing required"
	case BadValue:
		return "bad value"
	default:
		return "undefined form error type"
	}
}

const (
	MissingRequired = _FormErrorType(iota)
	BadValue
)

type FormError struct {
	FormName string
	Type     _FormErrorType
}

var CustomFormError func(fe *FormError) string

func init() {
	CustomFormError = func(fe *FormError) string {
		return fmt.Sprintf("FormError: %s; field `%s`", fe.Type, fe.FormName)
	}
}

func (e *FormError) Error() string {
	return CustomFormError(e)
}

func (e *FormError) StatusCode() int {
	return http.StatusBadRequest
}

func (rule *Rule) bindInterface(former Former, filed *reflect.Value) *FormError {
	var fv []byte
	var ok bool
	if len(rule.pathParamsName) > 0 {
		fv, ok = former.PathParam(rule.pathParamsName)
		if !ok {
			return &FormError{
				FormName: fmt.Sprintf("path param: %s", rule.pathParamsName),
				Type:     MissingRequired,
			}
		}
	} else {
		fv, ok = former.FormValue(rule.formName)
		if !ok {
			if rule.defaultVal != nil {
				filed.Set(reflect.ValueOf(rule.defaultVal))
				return nil
			} else {
				if rule.isRequired {
					return &FormError{FormName: internal.S(rule.formName), Type: MissingRequired}
				} else {
					return nil
				}
			}
		}
	}

	if !rule.notTrimSpace {
		fv = bytes.TrimSpace(fv)
	}

	var ret interface{}
	switch rule.rtype {
	case Bool:
		ret, ok = rule.toBool(fv)
	case Int64:
		ret, ok = rule.toInt(fv)
	case Uint64:
		ret, ok = rule.toUint(fv)
	case Float64:
		ret, ok = rule.toFloat(fv)
	case Bytes:
		ret, ok = rule.toBytes(fv)
	case String:
		ret, ok = rule.toString(fv)
	default:
		panic(fmt.Errorf("suna.validator: unexpected rule type"))
	}
	if ok {
		filed.Set(reflect.ValueOf(ret))
		return nil
	}
	return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
}

func (rule *Rule) bindSlice(former Former, field *reflect.Value) *FormError {
	var ret interface{}
	formVals := former.FormValues(rule.formName)
	if len(formVals) < 1 {
		if rule.isRequired {
			if rule.defaultVal != nil {
				field.Set(reflect.ValueOf(rule.defaultVal))
				return nil
			}
			return &FormError{FormName: internal.S(rule.formName), Type: MissingRequired}
		} else {
			return nil
		}
	}

	switch rule.rtype {
	case BoolSlice:
		var lst []bool
		for _, bs := range formVals {
			a, b := rule.toBool(bs)
			if !b {
				return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case IntSlice:
		var lst []int64
		for _, bs := range formVals {
			a, b := rule.toInt(bs)
			if !b {
				return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case UintSlice:
		var lst []uint64
		for _, bs := range formVals {
			a, b := rule.toUint(bs)
			if !b {
				return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case FloatSlice:
		var lst []float64
		for _, bs := range formVals {
			a, b := rule.toFloat(bs)
			if !b {
				return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case StringSlice:
		var lst []string
		for _, bs := range formVals {
			a, b := rule.toString(bs)
			if !b {
				return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case BytesSlice:
		var lst [][]byte
		for _, bs := range formVals {
			a, b := rule.toBytes(bs)
			if !b {
				return &FormError{FormName: internal.S(rule.formName), Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	default:
		panic(fmt.Errorf("suna.validator: unexpected rule type"))
	}

	v := reflect.ValueOf(ret)
	// check slice size
	if rule.fSSR {
		if v.IsNil() {
			if rule.isRequired {
				return &FormError{FormName: internal.S(rule.formName), Type: MissingRequired}
			}
			return nil
		}
		s := v.Len()
		if rule.minSSV != nil && s < *rule.minSSV {
			return &FormError{FormName: internal.S(rule.formName), Type: MissingRequired}
		}
		if rule.maxSSV != nil && s > *rule.maxSSV {
			return &FormError{FormName: internal.S(rule.formName), Type: MissingRequired}
		}
	}

	field.Set(v)
	return nil
}

func Validate(former Former, dist interface{}) (err *FormError, isNil bool) {
	v := reflect.ValueOf(dist).Elem()
	var field reflect.Value
	for _, rule := range GetRules(v.Type()) {
		field = v
		for _, index := range rule.fieldIndex {
			field = field.Field(index)
		}
		if rule.isSlice {
			err = rule.bindSlice(former, &field)
		} else {
			err = rule.bindInterface(former, &field)
		}
		if err != nil {
			return err, false
		}
	}
	return nil, true
}
