package validator

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/zzztttkkk/sha/internal"
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

var ErrPeekManyValuesFromURLParamsOrCookie = errors.New("sha.validator: peek many values from URLParams or cookie")

func fieldInfoToRule(t reflect.Type, f *reflectx.FieldInfo, defaultF func() interface{}) *Rule {
	rule := &Rule{
		fieldIndex: f.Index,
		fieldType:  f.Field.Type,
		formName:   f.Name,
		isRequired: true,
	}
	if len(rule.formName) < 1 {
		rule.formName = strings.ToLower(f.Field.Name)
	}

	if defaultF != nil {
		rule.defaultFunc = defaultF
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
			return nil
		}
	}

	for key, val := range f.Options {
		switch strings.ToLower(key) {
		case "w", "where":
			switch strings.ToLower(val) {
			case "query":
				rule.where = _WhereQuery
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.QueryValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { return former.QueryValues(name) }
			case "body":
				rule.where = _WhereBody
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.BodyValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { return former.BodyValues(name) }
			case "form":
				rule.where = _WhereForm
			case "url", "urlparams", "url-params", "urlparam", "url-param":
				rule.where = _WhereURLParams
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.URLParam(name) }
				rule.peekAll = func(former Former, name string) [][]byte { panic(ErrPeekManyValuesFromURLParamsOrCookie) }
			case "header":
				rule.where = _WhereHeader
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.HeaderValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { return former.HeaderValues(name) }
			case "cookie":
				rule.where = _WhereCookie
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.CookieValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { panic(ErrPeekManyValuesFromURLParamsOrCookie) }
			default:
				panic(
					fmt.Errorf(
						"sha.validator: bad where value, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		case "optional":
			rule.isRequired = false
		case "disabletrimspace", "disable-trim-space", "notrim", "no-trim":
			rule.notTrimSpace = true
		case "disableescapehtml", "disable-escape-html", "noescape", "no-escape":
			rule.notEscapeHtml = true
		case "description":
			rule.description = val
		case "r", "regexp":
			rule.reg = regexpMap[val]
			rule.regName = val
			if rule.reg == nil {
				panic(fmt.Errorf("sha.validator: unregistered regexp `%s`", val))
			}
		case "f", "filters", "filter":
			for _, n := range strings.Split(val, "|") {
				n = strings.TrimSpace(n)
				if len(n) < 1 {
					continue
				}
				f := bytesFilterMap[n]
				if f == nil {
					panic(fmt.Errorf("sha.validator: unregistered bytes filter `%s`", n))
				}
				rule.fns = append(rule.fns, f)
			}
			rule.fnNames = val
		case "l", "length", "len":
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
						"sha.validator: bad length range value, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		case "v", "value", "val":
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
			default:
				err = true
			}
			if err {
				panic(
					fmt.Errorf(
						"sha.validator: bad number value range or value range on non-numberic filed, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		case "s", "size":
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
						"sha.validator: bad slice size range, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		}
	}

	if rule.where == _WhereForm {
		rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.FormValue(name) }
		rule.peekAll = func(former Former, name string) [][]byte { return former.FormValues(name) }
	}
	return rule
}

func getDefaultFunc(ele reflect.Value, name string) func() interface{} {
	fnV := ele.MethodByName(fmt.Sprintf("Default%s", name))
	if fnV.IsValid() {
		ret, ok := fnV.Interface().(func() interface{})
		if ok {
			fmt.Println(fmt.Sprintf("Default%s", name))
			return ret
		}
	}
	return nil
}

func GetRules(t reflect.Type) Rules {
	v, ok := cacheMap[t]
	if ok {
		return v
	}

	ele := reflect.New(t)

	var rules Rules
	fmap := reflectMapper.TypeMap(t)
	for _, f := range fmap.Index {
		rule := fieldInfoToRule(t, f, getDefaultFunc(ele, f.Field.Name))
		if rule != nil {
			rules = append(rules, rule)
		}
	}
	cacheMap[t] = rules
	return rules
}
