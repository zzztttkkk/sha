package validator

import (
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sqlx/reflectx"
	"log"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

//CacheMap is not thread-safe, all types should be prepared before the server starts to listening.
var CacheMap = map[reflect.Type]Rules{}

const TagName = "vld"

var ReflectMapper = reflectx.NewMapper(TagName)

type Defaulter interface {
	Default(fieldName string) func() interface{}
}

type ExtDescriptor interface {
	ExtDescription() string
}

type DefaultAndExtDescriptor interface {
	Defaulter
	ExtDescriptor
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

func _isCustomField(rule *_Rule, t reflect.Type) {
	// type AField struct {}

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
		return
	}
}

var bytesType = reflect.TypeOf((*[]byte)(nil)).Elem()
var u8SliceType = reflect.TypeOf((*[]uint8)(nil)).Elem()

func isCustomField(rule *_Rule, t reflect.Type) {
	// bytes or struct
	if t == bytesType || t == u8SliceType || t.ConvertibleTo(bytesType) || t.Kind() != reflect.Slice {
		_isCustomField(rule, t)
		return
	}

	// slice of struct
	t = t.Elem()
	_isCustomField(rule, t)
	if rule.rtype == _CustomType {
		rule.isSlice = true
	}
}

func fieldInfoToRule(t reflect.Type, f *reflectx.FieldInfo, defaultF func() interface{}) *_Rule {
	rule := &_Rule{
		fieldIndex: f.Index,
		fieldType:  f.Field.Type,
		formName:   f.Name,
		isRequired: true,
	}

	if len(rule.formName) < 1 || (rule.formName == f.Field.Name && len(f.Options) == 0) {
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
			case "query", "q":
				rule.where = _WhereQuery
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.QueryValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { return former.QueryValues(name) }
			case "body", "b":
				rule.where = _WhereBody
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.BodyValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { return former.BodyValues(name) }
			case "form", "f":
				rule.where = _WhereForm
			case "url", "urlparams", "url-params", "urlparam", "url-param", "u":
				rule.where = _WhereURLParams
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.URLParam(name) }
				rule.peekAll = func(former Former, name string) [][]byte { panic(ErrPeekManyValuesFromURLParamsOrCookie) }
			case "header", "h", "headers":
				rule.where = _WhereHeader
				rule.peekOne = func(former Former, name string) ([]byte, bool) { return former.HeaderValue(name) }
				rule.peekAll = func(former Former, name string) [][]byte { return former.HeaderValues(name) }
			case "cookie", "c", "cookies":
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

// GetRules
/*
	tag name:
		vld
	tag options:
		where/w:
			- query/q => get bytes value from Request.Query
			- body/b => get bytes value from Request.Body
			- form/f => get bytes value from Request.Query or Request.Body
			- urlparams/url/u => get bytes value from Request.URLParams
			- headers/h => get bytes value from Request.Header
			- cookies/c => get bytes value from Request.Cookies

		optional
			this field is optional

		disabletrimspace/no-trim
			do not trim space of the bytes value

		disableescapehtml/no-escape
			do not escape of the string value

		description
			the description of this field

		regexp/r
			a name of the regexp
			use `RegisterRegexp` to register a regexp

		filter/f
			names of the bytes filter functions
			use `RegisterBytesFilter` to register a function

		value/v
			number value range

		size/s
			slice value size range

		length/len/l
			bytes length range
*/
func GetRules(t reflect.Type) Rules {
	v, ok := CacheMap[t]
	if ok {
		return v
	}

	ele := reflect.New(t)

	defaulter, isD := ele.Interface().(Defaulter)
	if !isD {
		defaulter, isD = ele.Elem().Interface().(Defaulter)
	}

	var rules Rules
	fMap := ReflectMapper.TypeMap(t)
	for _, f := range fMap.Index {
		var ed func() interface{}
		if isD {
			ed = defaulter.Default(f.Field.Name)
		}
		rule := fieldInfoToRule(t, f, ed)
		if rule != nil {
			rules = append(rules, rule)
		} else {
			log.Printf("sha.validator: ignore filed, `%s`.`%s`.`%s`", t.PkgPath(), t.Name(), f.Name)
		}
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].where < rules[j].where })

	CacheMap[t] = rules
	return rules
}
