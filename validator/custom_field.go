package validator

import (
	"reflect"
)

type Field interface {
	FromBytes(data []byte) error
	Validate() error
}

func (rule *_Rule) formValueToCustomField(f *reflect.Value, data []byte) error {
	t := rule.fieldType
	if rule.isPtr {
		t = t.Elem()
	}

	ptrV := reflect.New(t)
	err := rule.formValueToCustomFieldVPtr(&ptrV, data)
	if err != nil {
		return err
	}
	if rule.isPtr {
		f.Set(ptrV)
	} else {
		f.Set(ptrV.Elem())
	}
	return nil
}

func (rule *_Rule) formValueToCustomFieldVPtr(f *reflect.Value, data []byte) error {
	ptr := f.Interface().(Field)
	if err := ptr.FromBytes(data); err != nil {
		return &Error{FormName: rule.formName, Type: BadValue, Wrapped: err}
	}
	return nil
}

var fieldType = reflect.TypeOf((*Field)(nil)).Elem()

func toCustomField(v *reflect.Value) Field {
	if v.Type().ConvertibleTo(fieldType) {
		return v.Interface().(Field)
	}
	return v.Addr().Interface().(Field)
}
