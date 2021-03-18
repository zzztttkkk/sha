package validator

import (
	"github.com/zzztttkkk/sha/utils"
	"reflect"
)

func (rule *_Rule) validateOne(field *reflect.Value) *FormError {
	switch rule.rtype {
	case _Int64:
		if !rule.checkNumRange {
			return nil
		}
		i := field.Interface().(int64)
		if rule.minIntVal != nil && i < *rule.minIntVal {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxIntVal != nil && i > *rule.maxIntVal {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
	case _Float64:
		if !rule.checkNumRange {
			return nil
		}
		i := field.Interface().(float64)
		if rule.minDoubleVal != nil && i < *rule.minDoubleVal {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxDoubleVal != nil && i > *rule.maxDoubleVal {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
	case _Uint64:
		if !rule.checkNumRange {
			return nil
		}
		i := field.Interface().(uint64)
		if rule.minUintVal != nil && i < *rule.minUintVal {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxUintVal != nil && i > *rule.maxUintVal {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
	case _String:
		s, ok := rule.toString(utils.B(field.Interface().(string)))
		if !ok {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
		field.SetString(s)
	}
	return nil
}

func (rule *_Rule) validateSlice(field *reflect.Value) *FormError {
	vi := field.Interface()
	v := reflect.ValueOf(vi)
	if rule.checkListSize {
		i := v.Len()
		if rule.minSliceSize != nil && i < *rule.minSliceSize {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
		if rule.maxSliceSize != nil && 1 > *rule.maxSliceSize {
			return &FormError{FormName: rule.formName, Type: BadValue}
		}
	}

	switch rule.rtype {
	case _IntSlice:
		if rule.checkNumRange {
			for _, i := range vi.([]int64) {
				if rule.minIntVal != nil && i < *rule.minIntVal {
					return &FormError{FormName: rule.formName, Type: BadValue}
				}
				if rule.maxIntVal != nil && i < *rule.maxIntVal {
					return &FormError{FormName: rule.formName, Type: BadValue}
				}
			}
		}
	case _UintSlice:
		if rule.checkNumRange {
			for _, i := range vi.([]uint64) {
				if rule.minUintVal != nil && i < *rule.minUintVal {
					return &FormError{FormName: rule.formName, Type: BadValue}
				}
				if rule.maxUintVal != nil && i < *rule.maxUintVal {
					return &FormError{FormName: rule.formName, Type: BadValue}
				}
			}
		}
	case _FloatSlice:
		if rule.checkNumRange {
			for _, i := range vi.([]float64) {
				if rule.minDoubleVal != nil && i < *rule.minDoubleVal {
					return &FormError{FormName: rule.formName, Type: BadValue}
				}
				if rule.maxDoubleVal != nil && i < *rule.maxDoubleVal {
					return &FormError{FormName: rule.formName, Type: BadValue}
				}
			}
		}
	case _StringSlice:
		var ss []string
		for _, s := range vi.([]string) {
			_s, ok := rule.toString(utils.B(s))
			if !ok {
				return &FormError{FormName: rule.formName, Type: BadValue}
			}
			ss = append(ss, _s)
		}
		field.Set(reflect.ValueOf(ss))
	}
	return nil
}

func ValidateStruct(vPtr interface{}) (err *FormError) {
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
