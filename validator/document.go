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

func NewDocument(input interface{}, output interface{}) *SimpleDocument {
	o := ""
	if output != nil && reflect.ValueOf(output).IsValid() {
		s, _ := jsonx.Marshal(output)
		o = string(s)
	}
	var i string
	if input != nil && reflect.ValueOf(input).IsValid() {
		i = GetRules(reflect.TypeOf(input)).String()
		ed, ok := input.(ExtDescriptor)
		if ok && len(ed.ExtDescription()) > 0 {
			i += "\r\n`\r\n"
			i += ed.ExtDescription()
			i += "\r\n`"
		}
	}
	return &SimpleDocument{input: i, output: o}
}
