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

type _TagParser struct {
	current *_RuleT
	all     _RuleSliceT
	isJson  bool
}

func (p *_TagParser) OnNestedStruct(f *reflect.StructField) bool {
	return true
}

func (p *_TagParser) OnBegin(field *reflect.StructField) bool {
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
		p.all = append(p.all, subP.lst...)
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

func (p *_TagParser) OnName(name string) {
	p.current.form = name
}

func parseIntRange(s string) (int64, int64, bool, bool) {
	if len(s) < 1 {
		return 0, 0, false, false
	}

	if strings.HasPrefix(s, "-") {
		v, e := strconv.ParseInt(s[1:], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return 0, v, false, true
	}

	if strings.HasSuffix(s, "-") {
		v, e := strconv.ParseInt(s[:len(s)-1], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, 0, true, false
	}

	ss := strings.Split(s, "-")
	if len(ss) == 1 {
		v, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, v, true, true
	}

	if len(ss) != 2 {
		return 0, 0, false, false
	}

	minV, e := strconv.ParseInt(ss[0], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}

	maxV, e := strconv.ParseInt(ss[1], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}
	return minV, maxV, true, true
}

func parseUintRange(s string) (uint64, uint64, bool, bool) {
	if len(s) < 1 {
		return 0, 0, false, false
	}

	if strings.HasPrefix(s, "-") {
		v, e := strconv.ParseUint(s[1:], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return 0, v, false, true
	}

	if strings.HasSuffix(s, "-") {
		v, e := strconv.ParseUint(s[:len(s)-1], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, 0, true, false
	}

	ss := strings.Split(s, "-")
	if len(ss) == 1 {
		v, e := strconv.ParseUint(s, 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, v, true, true
	}
	if len(ss) != 2 {
		return 0, 0, false, false
	}

	minV, e := strconv.ParseUint(ss[0], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}

	maxV, e := strconv.ParseUint(ss[1], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}
	return minV, maxV, true, true
}

func (p *_TagParser) OnAttr(key, val string) {
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
		rule.minL, rule.maxL, rule.minLF, rule.maxLF = parseIntRange(val)
	case "V", "value":
		rule.vrange = true
		if rule.t == _Int64 {
			rule.minV, rule.maxV, rule.minVF, rule.maxVF = parseIntRange(val)
		} else if rule.t == _Uint64 {
			rule.minUV, rule.maxUV, rule.minUVF, rule.maxUVF = parseUintRange(val)
		}
	case "D", "default":
		rule.defaultV = []byte(val)
	case "S", "size":
		rule.srange = true
		rule.minS, rule.maxS, rule.minSF, rule.maxSF = parseIntRange(val)
	case "optional":
		rule.required = false
	}
}

func (p *_TagParser) OnDone() {
	p.all = append(p.all, p.current)
	p.current = nil
}

var rulesMap = map[reflect.Type]*Rules{}

func GetRules(v interface{}) *Rules { return getRules(reflect.TypeOf(v)) }

func getRules(p reflect.Type) *Rules {
	rs, ok := rulesMap[p]
	if ok {
		return rs
	}

	parser := &_TagParser{}
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

	rules := &Rules{
		lst:    parser.all,
		isJson: parser.isJson,
	}

	rulesMap[p] = rules
	return rules
}
