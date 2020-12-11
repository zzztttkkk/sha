package validator

import (
	"fmt"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/zzztttkkk/suna/internal"
	"log"
	"reflect"
	"strings"
)

// all validate type should be prepared before use
var cacheMap = map[reflect.Type]Rules{}
var reflectMapper = reflectx.NewMapper("validator")

type Defaulter interface {
	Default(fieldName string) interface{}
}

var customFieldType = reflect.TypeOf((*CustomField)(nil)).Elem()

func fieldInfoToRule(t reflect.Type, f *reflectx.FieldInfo, defaultF func(string2 string) interface{}) *Rule {
	rule := &Rule{
		fieldIndex: f.Index,
		fieldType:  f.Field.Type,
		formName:   []byte(f.Name),
		isRequired: true,
	}
	if len(rule.formName) < 1 {
		rule.formName = []byte(strings.ToLower(f.Field.Name))
	}

	if defaultF != nil {
		rule.defaultVal = defaultF(f.Field.Name)
	}

	elePtr := reflect.New(f.Field.Type)
	ele := elePtr.Elem()

	switch ele.Interface().(type) {
	case int64:
		rule.rtype = _Int64
	case uint64:
		rule.rtype = _Uint64
	case float64:
		rule.rtype = _Float64
	case bool:
		rule.rtype = _Bool
	case string:
		rule.rtype = _String
	case []byte:
		rule.rtype = _Bytes
	case [][]byte:
		rule.rtype = _BytesSlice
		rule.isSlice = true
	case []string:
		rule.rtype = _StringSlice
		rule.isSlice = true
	case []int64:
		rule.rtype = _IntSlice
		rule.isSlice = true
	case []uint64:
		rule.rtype = _UintSlice
		rule.isSlice = true
	case []bool:
		rule.rtype = _BoolSlice
		rule.isSlice = true
	case []float64:
		rule.rtype = _FloatSlice
		rule.isSlice = true
	case CustomField:
		rule.rtype = _CustomType
	default:
		if elePtr.Type().ConvertibleTo(customFieldType) {
			rule.rtype = _CustomType
			rule.indirectCustomField = true
		} else {
			log.Printf(
				"suna.validator: number is non-64bit value or unsupported type, `%s:%s.%s`\n",
				t.PkgPath(), t.Name(), f.Name,
			)
			return nil
		}
	}

	for key, val := range f.Options {
		switch key {
		case "P", "p", "params":
			rule.pathParamsName = []byte(val)
		case "optional":
			rule.isRequired = false
		case "NTSC", "nottrimspacechar":
			rule.notTrimSpace = true
		case "description":
			rule.description = val
		case "R", "r", "regexp":
			rule.reg = regexpMap[val]
			rule.regName = val
			if rule.reg == nil {
				panic(fmt.Errorf("suna.validator: unregistered regexp `%s`", val))
			}
		case "F", "f", "filter":
			rule.fn = bytesFilterMap[val]
			rule.fnName = val
			if rule.fn == nil {
				panic(fmt.Errorf("suna.validator: unregistered bytes filter `%s`", val))
			}
		case "L", "l", "length":
			rule.fLR = true
			minV, maxV, minVF, maxVF := internal.ParseIntRange(val)
			if minVF {
				rule.minLV = new(int)
				*rule.minLV = int(minV)
			}
			if maxVF {
				rule.maxLV = new(int)
				*rule.maxLV = int(maxV)
			}
			if rule.minLV == nil && rule.maxLV == nil {
				panic(
					fmt.Errorf(
						"suna.validator: bad length range value, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		case "V", "v", "value":
			rule.fVR = true
			var err bool
			switch rule.rtype {
			case _Int64, _IntSlice:
				minV, maxV, minVF, maxVF := internal.ParseIntRange(val)
				if minVF {
					rule.minIV = new(int64)
					*rule.minIV = minV
				}
				if maxVF {
					rule.maxIV = new(int64)
					*rule.maxIV = maxV
				}
				err = rule.minIV == nil && rule.maxIV == nil
			case _Uint64, _UintSlice:
				minV, maxV, minVF, maxVF := internal.ParseUintRange(val)
				if minVF {
					rule.minUV = new(uint64)
					*rule.minUV = minV
				}
				if maxVF {
					rule.maxUV = new(uint64)
					*rule.maxUV = maxV
				}
				err = rule.minUV == nil && rule.maxUV == nil
			case _Float64, _FloatSlice:
				minV, maxV, minVF, maxVF := internal.ParseFloatRange(val)
				if minVF {
					rule.minDV = new(float64)
					*rule.minDV = minV
				}
				if maxVF {
					rule.maxDV = new(float64)
					*rule.maxDV = maxV
				}
				err = rule.minDV == nil && rule.maxDV == nil
			}
			if err {
				panic(
					fmt.Errorf(
						"suna.validator: bad value range value, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		case "S", "s", "size":
			rule.fSSR = true
			minV, maxV, minVF, maxVF := internal.ParseIntRange(val)
			if minVF {
				rule.minSSV = new(int)
				*rule.minSSV = int(minV)
			}
			if maxVF {
				rule.maxSSV = new(int)
				*rule.maxSSV = int(maxV)
			}
			if rule.minSSV == nil && rule.maxSSV == nil {
				panic(
					fmt.Errorf(
						"suna.validator: bad slice size range value, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		}
	}

	return rule
}

func GetRules(t reflect.Type) Rules {
	v, ok := cacheMap[t]
	if ok {
		return v
	}

	ele := reflect.New(t).Elem().Interface()
	var d Defaulter
	var fn func(string) interface{}
	if d, ok = ele.(Defaulter); ok {
		fn = d.Default
	}

	var rules Rules
	fmap := reflectMapper.TypeMap(t)
	for _, f := range fmap.Index {
		rule := fieldInfoToRule(t, f, fn)
		if rule != nil {
			rules = append(rules, rule)
		}
	}
	cacheMap[t] = rules
	return rules
}
