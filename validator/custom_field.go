package validator

import (
	"reflect"
)

type Field interface {
	FromBytes(data []byte) error
	Validate() error
}

func (rule *_Rule) toCustomField(f *reflect.Value, data []byte) error {
	t := rule.fieldType
	if rule.isPtr {
		t = t.Elem()
	}

	ptrV := reflect.New(t)
	err := rule.toCustomFieldVPtr(&ptrV, data)
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

func (rule *_Rule) toCustomFieldVPtr(f *reflect.Value, data []byte) error {
	ptr := f.Interface().(Field)
	if err := ptr.FromBytes(data); err != nil {
		return err
	}
	if err := ptr.Validate(); err != nil {
		return err
	}
	return nil
}
