package validator

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"reflect"
	"unicode/utf8"
)

type Former interface {
	URLParam(name string) ([]byte, bool)

	QueryValue(name string) ([]byte, bool)
	QueryValues(name string) [][]byte

	BodyValue(name string) ([]byte, bool)
	BodyValues(name string) [][]byte

	FormValue(name string) ([]byte, bool)
	FormValues(name string) [][]byte

	HeaderValue(name string) ([]byte, bool)
	HeaderValues(name string) [][]byte
	CookieValue(name string) ([]byte, bool)
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

func htmlEscape(p []byte) []byte {
	var ret []byte
	var temp = make([]byte, 4)
	for _, r := range utils.S(p) {
		switch r {
		case '&':
			ret = append(ret, '&', 'a', 'm', 'p', ';')
		case '<':
			ret = append(ret, '&', 'l', 't', ';')
		case '>':
			ret = append(ret, '&', 'g', 't', ';')
		case '\'':
			ret = append(ret, '&', '#', '3', '9', ';')
		case '"':
			ret = append(ret, '&', '#', '3', '4', ';')
		default:
			l := utf8.EncodeRune(temp, r)
			for i := 0; i < l; i++ {
				ret = append(ret, temp[i])
			}
		}
	}
	return ret
}

func (rule *_Rule) bindOne(former Former, filed *reflect.Value) *Error {
	fv, ok := rule.peekOne(former, rule.formName)
	if !ok {
		if rule.where != _WhereURLParams {
			if rule.defaultFunc != nil {
				filed.Set(reflect.ValueOf(rule.defaultFunc()))
				return nil
			} else {
				if rule.isRequired {
					return &Error{FormName: rule.formName, Type: MissingRequired}
				} else {
					return nil
				}
			}
		} else {
			return &Error{FormName: rule.formName, Type: MissingRequired}
		}
	}

	var ret interface{}
	switch rule.rtype {
	case _Bool:
		ret, ok = rule.toBool(fv)
	case _Int64:
		ret, ok = rule.toInt(fv)
	case _Uint64:
		ret, ok = rule.toUint(fv)
	case _Float64:
		ret, ok = rule.toFloat(fv)
	case _Bytes:
		ret, ok = rule.toBytes(fv)
	case _String:
		ret, ok = rule.toString(fv)
	case _CustomType:
		var data []byte
		data, ok = rule.toBytes(fv)
		if ok {
			if err := rule.formValueToCustomField(filed, data); err != nil {
				return &Error{FormName: rule.formName, Type: BadValue, Wrapped: err}
			}
			return nil
		}
	default:
		panic(fmt.Errorf("sha.validator: unexpected rule type"))
	}
	if ok {
		if rule.isPtr {
			dist := reflect.New(rule.fieldType.Elem())
			dist.Elem().Set(reflect.ValueOf(ret))
			filed.Set(dist)
		} else {
			filed.Set(reflect.ValueOf(ret))
		}

		return nil
	}
	return &Error{FormName: rule.formName, Type: BadValue}
}

func (rule *_Rule) bindMany(former Former, field *reflect.Value) *Error {
	var ret interface{}
	formVals := rule.peekAll(former, rule.formName)
	if len(formVals) < 1 {
		if rule.isRequired {
			if rule.defaultFunc != nil {
				field.Set(reflect.ValueOf(rule.defaultFunc()))
				return nil
			}
			return &Error{FormName: rule.formName, Type: MissingRequired}
		} else {
			return nil
		}
	}

	switch rule.rtype {
	case _BoolSlice:
		var lst []bool
		for _, bs := range formVals {
			a, b := rule.toBool(bs)
			if !b {
				return &Error{FormName: rule.formName, Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case _IntSlice:
		var lst []int64
		for _, bs := range formVals {
			a, b := rule.toInt(bs)
			if !b {
				return &Error{FormName: rule.formName, Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case _UintSlice:
		var lst []uint64
		for _, bs := range formVals {
			a, b := rule.toUint(bs)
			if !b {
				return &Error{FormName: rule.formName, Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case _FloatSlice:
		var lst []float64
		for _, bs := range formVals {
			a, b := rule.toFloat(bs)
			if !b {
				return &Error{FormName: rule.formName, Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case _StringSlice:
		var lst []string
		for _, bs := range formVals {
			a, b := rule.toString(bs)
			if !b {
				return &Error{FormName: rule.formName, Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case _BytesSlice:
		var lst [][]byte
		for _, bs := range formVals {
			a, ok := rule.toBytes(bs)
			if !ok {
				return &Error{FormName: rule.formName, Type: BadValue}
			}
			lst = append(lst, a)
		}
		ret = lst
	case _CustomType:
		sliceV := reflect.MakeSlice(rule.fieldType, 0, len(formVals))
		eleT := rule.fieldType.Elem()
		if rule.isPtr {
			eleT = eleT.Elem()
		}

		for _, bs := range formVals {
			elePV := reflect.New(eleT)

			if err := rule.formValueToCustomFieldVPtr(&elePV, bs); err != nil {
				return &Error{FormName: rule.formName, Type: BadValue, Wrapped: err}
			}

			if rule.isPtr {
				sliceV = reflect.Append(sliceV, elePV)
			} else {
				sliceV = reflect.Append(sliceV, elePV.Elem())
			}
		}
		ret = sliceV.Interface()
	default:
		panic(fmt.Errorf("sha.validator: unexpected rule type"))
	}

	v := reflect.ValueOf(ret)
	// check slice size
	if rule.checkListSize {
		if v.IsNil() {
			if rule.isRequired {
				return &Error{FormName: rule.formName, Type: MissingRequired}
			}
			return nil
		}
		s := v.Len()
		if rule.minSliceSize != nil && s < *rule.minSliceSize {
			return &Error{FormName: rule.formName, Type: MissingRequired}
		}
		if rule.maxSliceSize != nil && s > *rule.maxSliceSize {
			return &Error{FormName: rule.formName, Type: MissingRequired}
		}
	}

	if rule.fieldType.Kind() == reflect.Ptr {
		dist := reflect.New(rule.fieldType.Elem())
		dist.Elem().Set(v)
		field.Set(dist)
	} else {
		field.Set(v)
	}
	return nil
}

// BindAndValidateForm return value is a ptr, not an interface.
func BindAndValidateForm(former Former, dist interface{}) (err *Error) {
	v := reflect.ValueOf(dist).Elem()
	var field reflect.Value
	for _, rule := range GetRules(v.Type()) {
		field = v
		for _, index := range rule.fieldIndex {
			field = field.Field(index)
		}
		if rule.isSlice {
			err = rule.bindMany(former, &field)
		} else {
			err = rule.bindOne(former, &field)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
