package internal

import (
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

func ReflectMap(t reflect.Type, fn func(*reflect.StructField), onStructField func(field *reflect.StructField) bool) {
	num := t.NumField()
	for i := 0; i < num; i++ {
		field := t.Field(i)

		if field.Type.Kind() == reflect.Struct {
			ok := true
			if onStructField != nil {
				ok = onStructField(&field)
			}
			if ok {
				ReflectMap(field.Type, fn, onStructField)
			}
			continue
		}

		fb := field.Name[0]
		if fb < 'A' || fb > 'Z' {
			continue
		}
		fn(&field)
	}
}

func ReflectKeys(rt reflect.Type, tag string, fn func(string) string) (lst []string) {
	ReflectMap(
		rt,
		func(field *reflect.StructField) {
			name := field.Tag.Get(tag)
			if len(name) < 1 {
				if fn != nil {
					name = fn(field.Name)
				} else {
					name = strings.ToLower(field.Name)
				}
			}
			if name == "-" {
				return
			}
			lst = append(lst, name)
		},
		func(field *reflect.StructField) bool {
			return field.Tag.Get(tag) != "-"
		},
	)
	return
}

var tagAttrReg = regexp.MustCompile(`^\w+<.*?>$`)

type TagParser interface {
	Tag() string
	OnField(f *reflect.StructField) bool
	OnName(name string)
	OnAttr(key, name string)
	OnDone()
}

// tag syntax
// [name:]AttrName[<AttrValue>;]...
func ReflectTags(p reflect.Type, parser TagParser) {
	tag := parser.Tag()
	ReflectMap(
		p,
		func(field *reflect.StructField) {
			t := field.Tag.Get(tag)
			if t == "-" {
				return
			}

			if !parser.OnField(field) {
				return
			}

			if t == "" {
				t = strings.ToLower(field.Name) + ":"
			} else if t[0] == ':' {
				t = strings.ToLower(field.Name) + t
			}

			ind := strings.IndexByte(t, ':')
			var name, attrs string
			if ind > -1 {
				name = t[:ind]
				attrs = t[ind+1:]

				if len(name) == 0 && len(attrs) == 0 {
					return
				}
			} else {
				panic("snow.internal.reflect: error tag")
			}

			parser.OnName(name)
			for _, v := range strings.Split(attrs, ";") {
				v = strings.TrimSpace(v)
				if len(v) == 0 {
					continue
				}

				var key, val string
				var err error
				if tagAttrReg.MatchString(v) {
					ind := strings.IndexByte(v, '<')
					val, err = url.QueryUnescape(strings.TrimSpace(v[ind+1 : len(v)-1]))
					if err != nil {
						panic(err)
					}
					key = strings.TrimSpace(v[:ind])
				} else {
					key = v
				}

				parser.OnAttr(key, val)
			}
			parser.OnDone()
		},
		func(field *reflect.StructField) bool {
			return field.Tag.Get(tag) != "-"
		},
	)
}
