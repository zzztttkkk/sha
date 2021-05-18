package validator

import (
	"github.com/zzztttkkk/sha/utils"
	"reflect"
)

func (rule *_Rule) validateOne(field *reflect.Value) *ValidateError {
	switch rule.rtype {
	case _CustomType:
		if err := toCustomField(field).Validate(); err != nil {
			return &ValidateError{FormName: rule.formName, Type: BadValue, Wrapped: err}
		}
		return nil
	case _Int64:
		if !rule.checkNumRange {
			return nil
		}
		i := field.Interface().(int64)
		if rule.minIntVal != nil && i < *rule.minIntVal {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxIntVal != nil && i > *rule.maxIntVal {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
	case _Float64:
		if !rule.checkNumRange {
			return nil
		}
		i := field.Interface().(float64)
		if rule.minDoubleVal != nil && i < *rule.minDoubleVal {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxDoubleVal != nil && i > *rule.maxDoubleVal {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
	case _Uint64:
		if !rule.checkNumRange {
			return nil
		}
		i := field.Interface().(uint64)
		if rule.minUintVal != nil && i < *rule.minUintVal {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxUintVal != nil && i > *rule.maxUintVal {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
	case _String:
		s, ok := rule.toString(utils.B(field.Interface().(string)))
		if !ok {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
		field.SetString(s)
	}
	return nil
}

func (rule *_Rule) validateSlice(field *reflect.Value) *ValidateError {
	vi := field.Interface()
	v := reflect.ValueOf(vi)
	if rule.checkListSize {
		i := v.Len()
		if rule.minSliceSize != nil && i < *rule.minSliceSize {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxSliceSize != nil && i > *rule.maxSliceSize {
			return &ValidateError{FormName: rule.formName, Type: BadValue}
		}
	}

	switch rule.rtype {
	case _IntSlice:
		if rule.checkNumRange {
			for _, i := range vi.([]int64) {
				if rule.minIntVal != nil && i < *rule.minIntVal {
					return &ValidateError{FormName: rule.formName, Type: BadValue}
				}
				if rule.maxIntVal != nil && i < *rule.maxIntVal {
					return &ValidateError{FormName: rule.formName, Type: BadValue}
				}
			}
		}
	case _UintSlice:
		if rule.checkNumRange {
			for _, i := range vi.([]uint64) {
				if rule.minUintVal != nil && i < *rule.minUintVal {
					return &ValidateError{FormName: rule.formName, Type: BadValue}
				}
				if rule.maxUintVal != nil && i < *rule.maxUintVal {
					return &ValidateError{FormName: rule.formName, Type: BadValue}
				}
			}
		}
	case _FloatSlice:
		if rule.checkNumRange {
			for _, i := range vi.([]float64) {
				if rule.minDoubleVal != nil && i < *rule.minDoubleVal {
					return &ValidateError{FormName: rule.formName, Type: BadValue}
				}
				if rule.maxDoubleVal != nil && i < *rule.maxDoubleVal {
					return &ValidateError{FormName: rule.formName, Type: BadValue}
				}
			}
		}
	case _StringSlice:
		var ss []string
		for _, s := range vi.([]string) {
			_s, ok := rule.toString(utils.B(s))
			if !ok {
				return &ValidateError{FormName: rule.formName, Type: BadValue}
			}
			ss = append(ss, _s)
		}
		field.Set(reflect.ValueOf(ss))
	case _CustomType:
		l := v.Len()
		for i := 0; i < l; i++ {
			ele := v.Index(i)
			if err := toCustomField(&ele).Validate(); err != nil {
				return &ValidateError{FormName: rule.formName, Type: BadValue, Wrapped: err}
			}
		}
	}
	return nil
}

func ValidateStruct(vPtr interface{}) (err *ValidateError) {
	v := reflect.ValueOf(vPtr).Elem()
	t := v.Type()
	for _, rule := range GetRules(t) {
		field := v
		for _, index := range rule.fieldIndex {
			field = field.Field(index)
		}
		if rule.isSlice {
			err = rule.validateSlice(&field)
		} else {
			err = rule.validateOne(&field)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
