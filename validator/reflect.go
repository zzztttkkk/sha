package validator

import (
	"fmt"
	"github.com/zzztttkkk/reflectx"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var funcMap = map[string]func([]byte) ([]byte, bool){}
var funcDescp = map[string]string{}

func RegisterRegexp(name string, reg *regexp.Regexp) {
	regexpMap[name] = reg
}

func RegisterFunc(name string, fn func([]byte) ([]byte, bool), descp string) {
	funcMap[name] = fn
	funcDescp[name] = descp
}

func init() {
	var space = regexp.MustCompile(`\s+`)
	var empty = []byte("")

	RegisterFunc(
		"username",
		func(i []byte) ([]byte, bool) {
			name := space.ReplaceAll(i, empty)
			count := utf8.RuneCount(name)
			return name, count >= 3 && count <= 20
		},
		"remove all space, and check rune count in range [3,20]",
	)
}

var rulesMap = map[reflect.Type]_RuleSliceT{}

type _ValidatorParser struct {
	current *_RuleT
	all     _RuleSliceT
}

func (p *_ValidatorParser) OnBegin(field *reflect.StructField) bool {
	rule := &_RuleT{field: field.Name, required: true}

	switch field.Type.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		rule.t = _Int64
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rule.t = _Uint64
	case reflect.Bool:
		rule.t = _Bool
	case reflect.String:
		rule.t = _String
	case reflect.Slice:
		switch field.Type.Elem().Kind() {
		case reflect.Uint8:
			rule.t = _Bytes
		case reflect.Bool:
			rule.t = _BoolSlice
		case reflect.Int64:
			rule.t = _IntSlice
		case reflect.Uint64:
			rule.t = _Uint64
		case reflect.String:
			rule.t = _StringSlice
		default:
			return false
		}
	case reflect.Struct:
		subRules := getRules(field.Type)
		p.all = append(p.all, subRules...)
		return false
	case reflect.Map:
		if field.Type.Key().Kind() == reflect.String && field.Type.Elem().Kind() == reflect.Interface {
			rule.t = _JsonObject
		} else {
			return false
		}
	default:
		return false
	}
	p.current = rule
	return true
}

func (p *_ValidatorParser) OnName(name string) {
	p.current.form = name
}

func (p *_ValidatorParser) OnAttr(key, val string) {
	var rangeErr = fmt.Errorf("snow.validator: error length range `%s`", val)
	rule := p.current
	switch key {
	case "R", "regexp":
		rule.reg = regexpMap[val]
		rule.regName = val
		if rule.reg == nil {
			panic(fmt.Errorf("snow.validator: unregistered regexp `%s`", val))
		}
	case "F", "function":
		rule.fn = funcMap[val]
		rule.fnName = val
		if rule.fn == nil {
			panic(fmt.Errorf("snow.validator: unregistered function `%s`", val))
		}
	case "L", "length":
		rule.lrange = true
		vs := strings.Split(val, "-")
		if len(vs) != 2 {
			panic(rangeErr)
		}
		minL, e := strconv.ParseInt(vs[0], 10, 32)
		if e != nil {
			panic(rangeErr)
		}
		rule.minL = int(minL)
		maxL, e := strconv.ParseInt(vs[1], 10, 32)
		if e != nil {
			panic(rangeErr)
		}
		rule.maxL = int(maxL)
	case "V", "value":
		rule.vrange = true
		vs := strings.Split(val, "-")
		if len(vs) != 2 {
			panic(rangeErr)
		}
		if rule.t == _Int64 {
			minV, e := strconv.ParseInt(vs[0], 10, 64)
			if e != nil {
				panic(rangeErr)
			}
			rule.minV = minV
			maxV, e := strconv.ParseInt(vs[1], 10, 64)
			if e != nil {
				panic(rangeErr)
			}
			rule.maxV = maxV
		} else if rule.t == _Uint64 {
			minV, e := strconv.ParseUint(vs[0], 10, 64)
			if e != nil {
				panic(rangeErr)
			}
			rule.minUV = minV
			maxV, e := strconv.ParseUint(vs[1], 10, 64)
			if e != nil {
				panic(rangeErr)
			}
			rule.maxUV = maxV
		}
	case "D", "default":
		rule.defaultV = []byte(val)
	case "optional":
		rule.required = false
	}
}

func (p *_ValidatorParser) OnDone() {
	p.all = append(p.all, p.current)
	p.current = nil
}

func getRules(p reflect.Type) _RuleSliceT {
	rs, ok := rulesMap[p]
	if ok {
		return rs
	}

	parser := _ValidatorParser{}
	reflectx.Tags(p, "validator", &parser)
	sort.Sort(parser.all)

	rulesMap[p] = parser.all
	return parser.all
}
