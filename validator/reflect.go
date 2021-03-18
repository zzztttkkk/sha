package validator

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/zzztttkkk/sha/internal"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

var cacheMap = map[reflect.Type]Rules{}
var reflectMapper = reflectx.NewMapper("validator")

type Defaulter interface {
	Default(fieldName string) interface{}
}

var customFieldType = reflect.TypeOf((*Field)(nil)).Elem()

var ErrPeekManyValuesFromURLParamsOrCookie = errors.New("sha.validator: peek many values from URLParams or cookie")

func setType(rule *_Rule, t reflect.Type) bool {
	elePtr := reflect.New(t)
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
	default:
		return false
	}
	return true
}

var NameCast func(fieldName string) string

func init() {
	NameCast = func(fieldName string) string {
		var sb strings.Builder
		doSep := false

		for _, r := range fieldName {
			if unicode.IsUpper(r) {
				if doSep {
					sb.WriteByte('_')
					doSep = false
				}
				sb.WriteRune(unicode.ToLower(r))
			} else {
				doSep = true
				sb.WriteRune(r)
			}
		}
		return sb.String()
	}
}

func isCustomField(rule *_Rule, t reflect.Type) {
	// type AField **

	if t.Kind() == reflect.Ptr { // *AField
		if t.ConvertibleTo(customFieldType) {
			rule.isPtr = true
			rule.rtype = _CustomType
		}
		return
	}

	ele := reflect.New(t)
	if ele.Type().ConvertibleTo(customFieldType) {
		rule.rtype = _CustomType
	}
}

func fieldInfoToRule(t reflect.Type, f *reflectx.FieldInfo, defaultF func() interface{}) *_Rule {
	rule := &_Rule{
		fieldIndex: f.Index,
		fieldType:  f.Field.Type,
		formName:   f.Name,
		isRequired: true,
	}

	if len(rule.formName) < 1 || len(f.Options) == 0 {
		rule.formName = NameCast(f.Field.Name)
	}

	if defaultF != nil {
		rule.defaultFunc = defaultF
	}

	ft := f.Field.Type
	isCustomField(rule, ft)
	if rule.rtype != _CustomType {
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
			rule.isPtr = true
		}
		if !setType(rule, ft) {
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
		case "v", "value", "val":
			rule.checkNumRange = true
			var err bool
			switch rule.rtype {
			case _Int64, _IntSlice:
				minV, maxV, minVF, maxVF := internal.ParseIntRange(val)
				if minVF {
					rule.minIntVal = new(int64)
					*rule.minIntVal = minV
				}
				if maxVF {
					rule.maxIntVal = new(int64)
					*rule.maxIntVal = maxV
				}
				err = rule.minIntVal == nil && rule.maxIntVal == nil
			case _Uint64, _UintSlice:
				minV, maxV, minVF, maxVF := internal.ParseUintRange(val)
				if minVF {
					rule.minUintVal = new(uint64)
					*rule.minUintVal = minV
				}
				if maxVF {
					rule.maxUintVal = new(uint64)
					*rule.maxUintVal = maxV
				}
				err = rule.minUintVal == nil && rule.maxUintVal == nil
			case _Float64, _FloatSlice:
				minV, maxV, minVF, maxVF := internal.ParseFloatRange(val)
				if minVF {
					rule.minDoubleVal = new(float64)
					*rule.minDoubleVal = minV
				}
				if maxVF {
					rule.maxDoubleVal = new(float64)
					*rule.maxDoubleVal = maxV
				}
				err = rule.minDoubleVal == nil && rule.maxDoubleVal == nil
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
			rule.checkListSize = true
			minV, maxV, minVF, maxVF := internal.ParseIntRange(val)
			if minVF {
				rule.minSliceSize = new(int)
				*rule.minSliceSize = int(minV)
			}
			if maxVF {
				rule.maxSliceSize = new(int)
				*rule.maxSliceSize = int(maxV)
			}
			if rule.minSliceSize == nil && rule.maxSliceSize == nil {
				panic(
					fmt.Errorf(
						"sha.validator: bad slice size range, field: `%s:%s.%s`, tag value: `%s`",
						t.PkgPath(), t.Name(), f.Field.Name, val,
					),
				)
			}
		case "l", "length", "len":
			rule.checkFieldBytesSize = true
			minV, maxV, minVF, maxVF := internal.ParseIntRange(val)
			if minVF {
				rule.minFieldBytesSize = new(int)
				*rule.minFieldBytesSize = int(minV)
			}
			if maxVF {
				rule.maxFieldBytesSize = new(int)
				*rule.maxFieldBytesSize = int(maxV)
			}
			if rule.minFieldBytesSize == nil && rule.maxFieldBytesSize == nil {
				panic(
					fmt.Errorf(
						"sha.validator: bad field bytes size range, field: `%s:%s.%s`, tag value: `%s`",
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
	sort.Sort(rules)

	cacheMap[t] = rules
	return rules
}
