package validator

import (
	"fmt"
	"github.com/zzztttkkk/suna/reflectx"
	"github.com/zzztttkkk/suna/utils"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type _ValidatorParser struct {
	current *_RuleT
	all     _RuleSliceT
	isJson  bool
}

func (p *_ValidatorParser) OnNestStruct(f *reflect.StructField) bool {
	return f.Anonymous
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
		rule.isSlice = true
		ele := reflect.MakeSlice(field.Type, 1, 1).Interface()
		switch ele.(type) {
		case [][]byte:
			rule.t = _BytesSlice
		case []uint8:
			rule.t = _Bytes
			rule.isSlice = false
		case []int64:
			rule.t = _IntSlice
		case []uint64:
			rule.t = _UintSlice
		case []bool:
			rule.t = _BoolSlice
		case []string:
			rule.t = _StringSlice
		case utils.JsonArray:
			rule.isSlice = false
			rule.t = _JsonArray
		case JoinedBoolSlice:
			rule.t = _JoinedBoolSlice
			rule.isJoined = true
		case JoinedIntSlice:
			rule.t = _JoinedIntSlice
			rule.isJoined = true
		case JoinedUintSlice:
			rule.t = _JoinedUintSlice
			rule.isJoined = true
		default:
			return false
		}
	case reflect.Struct:
		subP := GetRules(field.Type)
		p.all = append(p.all, subP.all...)
		return false
	case reflect.Map:
		ele := reflect.MakeMap(field.Type).Interface()
		switch ele.(type) {
		case utils.JsonObject:
			rule.t = _JsonObject
		default:
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
	var rangeErr = fmt.Errorf("suna.validator: error range `%s`", val)
	rule := p.current
	switch key {
	case "R", "regexp":
		rule.reg = regexpMap[val]
		rule.regName = val
		if rule.reg == nil {
			panic(fmt.Errorf("suna.validator: unregistered regexp `%s`", val))
		}
	case "F", "function":
		rule.fn = funcMap[val]
		rule.fnName = val
		if rule.fn == nil {
			panic(fmt.Errorf("suna.validator: unregistered function `%s`", val))
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
		rule.minL = minL
		maxL, e := strconv.ParseInt(vs[1], 10, 32)
		if e != nil {
			panic(rangeErr)
		}
		rule.maxL = maxL

		if rule.maxL < rule.minL {
			panic(rangeErr)
		}
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
			if rule.maxV < rule.minV {
				panic(rangeErr)
			}
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
			if rule.maxUV < rule.minUV {
				panic(rangeErr)
			}
		}
	case "D", "default":
		rule.defaultV = []byte(val)
	case "S", "size":
		rule.srange = true
		vs := strings.Split(val, "-")
		if len(vs) != 2 {
			panic(rangeErr)
		}
		minV, e := strconv.ParseInt(vs[0], 10, 64)
		if e != nil {
			panic(rangeErr)
		}
		rule.minS = minV
		maxV, e := strconv.ParseInt(vs[1], 10, 64)
		if e != nil {
			panic(rangeErr)
		}
		rule.maxS = maxV
		if rule.maxS < rule.minS {
			panic(rangeErr)
		}
	case "optional":
		rule.required = false
	}
}

func (p *_ValidatorParser) OnDone() {
	p.all = append(p.all, p.current)
	p.current = nil
}

var rulesMap = map[reflect.Type]*_ValidatorParser{}

func GetRules(p reflect.Type) *_ValidatorParser {
	rs, ok := rulesMap[p]
	if ok {
		return rs
	}

	parser := &_ValidatorParser{}
	reflectx.Tags(p, "validator", parser)
	sort.Sort(parser.all)

	for _, r := range parser.all {
		if r.t == _JsonObject || r.t == _JsonArray {
			parser.isJson = true
		}
	}

	if parser.isJson && len(parser.all) != 1 {
		log.Fatalf("suna.validator: %s is contain a json field", p.Name())
	}

	rulesMap[p] = parser
	return parser
}
