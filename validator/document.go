package validator

import (
	"github.com/zzztttkkk/sha/jsonx"
	"reflect"
)

type Document interface {
	Tags() []string
	Description() string
	Input() string
	Output() string
}

type SimpleDocument struct {
	description string
	input       string
	output      string
	tags        []string
}

var _ Document = (*SimpleDocument)(nil)

func (m *SimpleDocument) Input() string {
	return m.input
}

func (m *SimpleDocument) Output() string {
	return m.output
}

func (m *SimpleDocument) Tags() []string {
	return m.tags
}

func (m *SimpleDocument) AddTags(tags ...string) *SimpleDocument {
	m.tags = append(m.tags, tags...)
	return m
}

func (m *SimpleDocument) Description() string {
	return m.description
}

func (m *SimpleDocument) SetDescription(desc string) *SimpleDocument {
	m.description = desc
	return m
}

type _undefined struct{}

var Undefined = &_undefined{}

func NewDocument(input interface{}, output interface{}) *SimpleDocument {
	o := ""
	if output == Undefined {
		o = "Undefined"
	} else {
		if output == nil || !reflect.ValueOf(output).IsValid() {
			o = "Empty"
		} else {
			s, _ := jsonx.Marshal(output)
			o = string(s)
		}
	}

	var i string
	if input == Undefined {
		i = "Undefined"
	} else {
		if input == nil || !reflect.ValueOf(input).IsValid() {
			i = "Empty"
		} else {
			i = GetRules(reflect.TypeOf(input)).String()
		}
	}
	return &SimpleDocument{input: i, output: o}
}
