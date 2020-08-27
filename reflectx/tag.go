package reflectx

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

type TagParser interface {
	OnNestedStruct(f *reflect.StructField) bool
	OnBegin(f *reflect.StructField) bool
	OnName(name string)
	OnAttr(key, name string)
	OnDone()
}

var tagAttrReg = regexp.MustCompile(`^(?P<name>\w+)(<(?P<value>.*?)>)?$`)

func parseAttr(item string) (name, value string, ok bool) {
	match := tagAttrReg.FindStringSubmatch(item)
	if len(match) < 2 {
		return "", "", false
	}

	m := make(map[string]string)
	for i, name := range tagAttrReg.SubexpNames() {
		if i != 0 {
			m[name] = match[i]
		}
	}
	return m["name"], m["value"], true
}

func parseTag(p reflect.Type, tag string, parser TagParser, field *reflect.StructField) {
	value := field.Tag.Get(tag)
	if value == "-" || !parser.OnBegin(field) {
		return
	}

	ind := strings.IndexByte(value, ':')
	var name, attrs string
	if ind > -1 {
		name = value[:ind]
		attrs = value[ind+1:]
	} else {
		attrs = value
	}

	if len(name) < 1 {
		name = strings.ToLower(field.Name)
	}
	parser.OnName(name)

	for _, item := range strings.Split(attrs, ";") {
		item = strings.TrimSpace(item)
		if len(item) == 0 {
			continue
		}

		key, val, ok := parseAttr(item)
		if ok {
			val, err := url.QueryUnescape(val)
			if err != nil {
				panic(
					fmt.Errorf(
						"reflectx: field `%s`.`%s`, tag `%s`", p.Name(), field.Name, tag,
					),
				)
			}
			parser.OnAttr(key, val)
		}
	}

	parser.OnDone()
}

// tag syntax
// [currentName:]AttrName[<AttrValue>;]...
func Tags(p reflect.Type, tag string, parser TagParser) {
	Map(
		p,
		&_VisitorT{
			onNS: func(field *reflect.StructField) bool {
				t := field.Tag.Get(tag)
				if t == "-" {
					return false
				}
				if parser.OnNestedStruct(field) {
					return true
				}
				parseTag(p, tag, parser, field)
				return false
			},
			onF: func(field *reflect.StructField) {
				parseTag(p, tag, parser, field)
			},
		},
	)
}
