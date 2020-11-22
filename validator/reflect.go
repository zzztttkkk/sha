package validator

import (
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/internal/typereflect"
	"log"
	"reflect"
)

type _TagParser struct {
	current *Rule
	field   *reflect.StructField
	rules   Rules
	name    string
}

func (p *_TagParser) OnNestedStruct(f *reflect.StructField, index []int) typereflect.OnNestStructRet {
	if !f.Anonymous {
		log.Println(
			fmt.Sprintf(
				"suna.validator: skip nested-struct `%s.%s`, because it is not anonymous",
				p.name, f.Name,
			),
		)
		return typereflect.Skip
	}
	return typereflect.GoDown
}

func (p *_TagParser) OnBegin(f *reflect.StructField, index []int) bool {
	p.field = f

	if f.Type.Kind() == reflect.Struct {
		p.rules = append(p.rules, GetRules(f.Type)...)
		return false
	}
	if f.Type.Kind() == reflect.Ptr {
		log.Println("suna.validator: field is isRequired by default, do not use pointer")
		return false
	}

	p.current = &Rule{}
	rule := p.current

	rule.fieldIndex = append(p.current.fieldIndex, index...)
	rule.isRequired = true

	ele := reflect.New(f.Type).Elem()
	switch ele.Interface().(type) {
	case int64:
		rule.rtype = Int64
	case uint64:
		rule.rtype = Uint64
	case float64:
		rule.rtype = Float64
	case bool:
		rule.rtype = Bool
	case string:
		rule.rtype = String
	case []byte:
		rule.rtype = Bytes
	case [][]byte:
		rule.rtype = BytesSlice
		rule.isSlice = true
	case []string:
		rule.rtype = StringSlice
		rule.isSlice = true
	case []int64:
		rule.rtype = IntSlice
		rule.isSlice = true
	case []uint64:
		rule.rtype = UintSlice
		rule.isSlice = true
	case []bool:
		rule.rtype = BoolSlice
		rule.isSlice = true
	case []float64:
		rule.rtype = FloatSlice
		rule.isSlice = true
	default:
		log.Printf("suna.validator: number is non-64bit value or unsupported type, `%s.%s`", p.name, f.Name)
		p.current = nil
		return false
	}

	return true
}

func (p *_TagParser) OnName(name string) {
	p.current.formName = []byte(name)
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
	case "F", "filter":
		rule.fn = bytesFilterMap[val]
		rule.fnName = val
		if rule.fn == nil {
			panic(fmt.Errorf("suna.validator: unregistered bytes filter `%s`", val))
		}
	case "L", "length":
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
					"suna.validator: bad field length range value, field: `%s.%s`, tag: `%s`",
					p.name, p.field.Name, val,
				),
			)
		}
	case "V", "value":
		rule.fVR = true
		var err bool
		switch rule.rtype {
		case Int64, IntSlice:
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
		case Uint64, UintSlice:
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
		case Float64, FloatSlice:
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
					"suna.validator: bad value range value, field: `%s.%s`, tag: `%s`",
					p.name, p.field.Name, val,
				),
			)
		}
	case "D", "default":
		if rule.isSlice {
			log.Println(fmt.Sprintf("suna.validator: ignore default value on a slice field, `%s.%s`", p.name, p.field.Name))
		}

		v := new([]byte)
		*v = append(*v, []byte(val)...)
		rule.defaultValPtr = v
	case "S", "size":
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
					"suna.validator: bad slice size range value, field: `%s.%s`, tag: `%s`",
					p.name, p.field.Name, val,
				),
			)
		}
	case "P", "params":
		rule.pathParamsName = []byte(val)
	case "optional":
		rule.isRequired = false
	case "NTSC", "nottrimspacechar":
		rule.notTrimSpace = true
	case "description":
		rule.description = val
	}
}

func (p *_TagParser) OnDone() {
	rule := p.current
	p.current = nil
	p.field = nil

	if !rule.fLR { // field can not be empty by default
		rule.fLR = true
		rule.minLV = new(int)
		*rule.minLV = 1
	}

	if rule.isSlice && !rule.fSSR {
		rule.fSSR = true
		rule.minSSV = new(int)
		*rule.minSSV = 1
	}
	p.rules = append(p.rules, rule)
}

// all validate type should be prepared before use
var cacheMap = map[reflect.Type]Rules{}

func GetRules(t reflect.Type) Rules {
	v, ok := cacheMap[t]
	if ok {
		return v
	}

	parser := _TagParser{name: fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())}
	typereflect.Tags(t, "validator", &parser)

	cacheMap[t] = parser.rules
	return parser.rules
}
