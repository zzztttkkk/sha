package validator

import "reflect"

type CustomField interface {
	FormValue(data []byte) bool
}

func (rule *Rule) toCustomField(data []byte) (interface{}, bool) {
	if rule.indirectCustomField {
		ele := reflect.New(rule.fieldType)
		ok := ele.Interface().(CustomField).FormValue(data)
		return ele.Elem().Interface(), ok
	}

	ele := reflect.New(rule.fieldType.Elem()).Interface()
	ok := ele.(CustomField).FormValue(data)
	return ele, ok
}
