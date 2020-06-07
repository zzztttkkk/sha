package validator

import (
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/internal"
	"github.com/zzztttkkk/snow/output"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var funcMap = map[string]func([]byte) ([]byte, bool){}

func RegisterRegexp(name string, reg *regexp.Regexp) {
	regexpMap[name] = reg
}

func RegisterFunc(name string, fn func([]byte) ([]byte, bool)) {
	funcMap[name] = fn
}

func init() {
	var space = regexp.MustCompile(`\s+`)
	var empty = []byte("")
	funcMap["username"] = func(i []byte) ([]byte, bool) { return space.ReplaceAll(i, empty), true }
}

var rulesMap = map[reflect.Type]_RuleSliceT{}

type _ValidatorParser struct {
	current *_RuleT
	all     _RuleSliceT
}

func (p *_ValidatorParser) Tag() string {
	return "vld"
}

func (p *_ValidatorParser) OnField(field *reflect.StructField) bool {
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
		if field.Type.Elem().Kind() == reflect.Uint8 {
			rule.t = _Bytes
		} else {
			return false
		}
	case reflect.Struct:
		subRules := getRules(field.Type)
		p.all = append(p.all, subRules...)
		return false
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
	internal.ReflectTags(p, &parser)
	sort.Sort(parser.all)
	rulesMap[p] = parser.all
	return parser.all
}

func Validate(ctx *fasthttp.RequestCtx, ptr interface{}) bool {
	_v := reflect.ValueOf(ptr).Elem()
	for _, rule := range getRules(_v.Type()) {
		val := ctx.FormValue(rule.form)
		if val != nil {
			val = bytes.TrimSpace(val)
		}

		if len(val) == 0 && len(rule.defaultV) > 0 {
			val = rule.defaultV
		}

		if len(val) == 0 {
			if rule.required {
				output.StdError(ctx, fasthttp.StatusBadRequest)
				return false
			}
			continue
		}

		field := _v.FieldByName(rule.field)
		switch rule.t {
		case _Bool:
			field.SetBool(rule.toBool(val))
		case _Int64:
			v, ok := rule.toI64(val)
			if !ok {
				output.StdError(ctx, fasthttp.StatusBadRequest)
				return false
			}
			field.SetInt(v)
		case _Uint64:
			v, ok := rule.toUI64(val)
			if !ok {
				output.StdError(ctx, fasthttp.StatusBadRequest)
				return false
			}
			field.SetUint(v)
		case _Bytes:
			v, ok := rule.toBytes(val)
			if !ok {
				output.StdError(ctx, fasthttp.StatusBadRequest)
				return false
			}
			field.SetBytes(v)
		case _String:
			v, ok := rule.toBytes(val)
			if !ok {
				output.StdError(ctx, fasthttp.StatusBadRequest)
				return false
			}
			field.SetString(internal.B2s(v))
		}
	}

	return true
}
