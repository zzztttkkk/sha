package validator

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"html"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type _RuleType int

const (
	_Bool = _RuleType(iota)
	_Int64
	_Uint64
	_Float64
	_Bytes
	_String

	_BoolSlice
	_IntSlice
	_UintSlice
	_FloatSlice
	_StringSlice
	_BytesSlice

	_CustomType
)

var typeNames = []string{
	"Bool",
	"Int",
	"Uint",
	"Float",
	"String",
	"String",

	"BoolArray",
	"IntArray",
	"UintArray",
	"FloatArray",
	"StringArray",
	"StringArray",

	"CustomType",
}

const (
	_WhereForm = iota
	_WhereURLParams
	_WhereQuery
	_WhereBody
	_WhereCookie
	_WhereHeader
)

type _Rule struct {
	fieldIndex []int
	fieldType  reflect.Type
	isPtr      bool

	// where
	formName string
	rtype    _RuleType
	where    int
	peekOne  func(former Former, name string) ([]byte, bool)
	peekAll  func(former Former, name string) [][]byte

	isRequired  bool
	description string

	checkNumRange bool // int value range flag
	minIntVal     *int64
	maxIntVal     *int64
	minUintVal    *uint64
	maxUintVal    *uint64
	minDoubleVal  *float64
	maxDoubleVal  *float64

	isSlice       bool
	checkListSize bool // slice size range flag
	minSliceSize  *int
	maxSliceSize  *int

	checkFieldBytesSize bool
	minFieldBytesSize   *int
	maxFieldBytesSize   *int

	notEscapeHtml bool
	notTrimSpace  bool

	defaultFunc func() interface{}

	reg     *regexp.Regexp
	regName string

	fns     []func([]byte) ([]byte, bool)
	fnNames string
}

var MarkdownTableHeader = "\n|name|type|required|field bytes size|int value range|list size range|default|regexp|function|description|\n"

func init() {
	MarkdownTableHeader += strings.Repeat("|:---", 10)
	MarkdownTableHeader += "|\n"
}

var ruleFmt = utils.NewNamedFmt(
	"|${name}|${type}|${required}|${lrange}|${vrange}|${srange}|${default}|${regexp}|${function}|${description}|",
)

// markdown table row
func (rule *_Rule) String() string {
	typeString := typeNames[rule.rtype]
	if rule.rtype == _CustomType {
		if rule.isSlice {
			eleT := rule.fieldType.Elem()
			if eleT.Kind() == reflect.Ptr {
				typeString = fmt.Sprintf("[]%s", reflect.New(eleT.Elem()).Type().Elem().Name())
			} else {
				typeString = fmt.Sprintf("[]%s", eleT.Name())
			}
		} else {
			if rule.isPtr {
				typeString = reflect.New(rule.fieldType).Type().Elem().Name()
			} else {
				typeString = rule.fieldType.Name()
			}
		}
	}
	m := utils.M{
		"type":     typeString,
		"required": fmt.Sprintf("%v", rule.isRequired),
	}

	if len(rule.description) > 0 {
		m["description"] = rule.description
	} else {
		m["description"] = "/"
	}

	switch rule.where {
	case _WhereForm:
		m["name"] = fmt.Sprintf("Form{%s}", rule.formName)
	case _WhereCookie:
		m["name"] = fmt.Sprintf("Cookie{%s}", rule.formName)
	case _WhereHeader:
		m["name"] = fmt.Sprintf("Header{%s}", rule.formName)
	case _WhereBody:
		m["name"] = fmt.Sprintf("Body{%s}", rule.formName)
	case _WhereQuery:
		m["name"] = fmt.Sprintf("Query{%s}", rule.formName)
	case _WhereURLParams:
		m["name"] = fmt.Sprintf("URLParams{%s}", rule.formName)
	}

	if rule.checkListSize {
		if rule.minSliceSize != nil && rule.maxSliceSize != nil {
			m["srange"] = fmt.Sprintf("%d-%d", *rule.minSliceSize, *rule.maxSliceSize)
		} else if rule.minSliceSize != nil {
			m["srange"] = fmt.Sprintf("%d-", *rule.minSliceSize)
		} else if rule.maxSliceSize != nil {
			m["srange"] = fmt.Sprintf("-%d", *rule.maxSliceSize)
		} else {
			m["srange"] = "/"
		}
	} else {
		m["srange"] = "/"
	}

	if rule.checkFieldBytesSize {
		if rule.minFieldBytesSize != nil && rule.maxFieldBytesSize != nil {
			m["lrange"] = fmt.Sprintf("%d-%d", *rule.minFieldBytesSize, *rule.maxFieldBytesSize)
		} else if rule.minFieldBytesSize != nil {
			m["lrange"] = fmt.Sprintf("%d-", *rule.minFieldBytesSize)
		} else if rule.maxFieldBytesSize != nil {
			m["lrange"] = fmt.Sprintf("-%d", *rule.maxFieldBytesSize)
		} else {
			m["lrange"] = "/"
		}
	} else {
		m["lrange"] = "/"
	}

	if rule.checkNumRange {
		switch rule.rtype {
		case _Int64, _IntSlice:
			if rule.minIntVal != nil && rule.maxIntVal != nil {
				m["vrange"] = fmt.Sprintf("%d-%d", *rule.minIntVal, *rule.maxIntVal)
			} else if rule.minIntVal != nil {
				m["vrange"] = fmt.Sprintf("%d-", *rule.minIntVal)
			} else if rule.maxIntVal != nil {
				m["vrange"] = fmt.Sprintf("-%d", *rule.maxIntVal)
			} else {
				m["vrange"] = "/"
			}
		case _Uint64, _UintSlice:
			if rule.minUintVal != nil && rule.maxUintVal != nil {
				m["vrange"] = fmt.Sprintf("%d-%d", *rule.minUintVal, *rule.maxUintVal)
			} else if rule.minUintVal != nil {
				m["vrange"] = fmt.Sprintf("%d-", *rule.minUintVal)
			} else if rule.maxUintVal != nil {
				m["vrange"] = fmt.Sprintf("-%d", *rule.maxUintVal)
			} else {
				m["vrange"] = "/"
			}
		case _Float64, _FloatSlice:
			if rule.minDoubleVal != nil && rule.maxDoubleVal != nil {
				m["vrange"] = fmt.Sprintf("%f-%f", *rule.minDoubleVal, *rule.maxDoubleVal)
			} else if rule.minDoubleVal != nil {
				m["vrange"] = fmt.Sprintf("%f-", *rule.minDoubleVal)
			} else if rule.maxDoubleVal != nil {
				m["vrange"] = fmt.Sprintf("-%f", *rule.maxDoubleVal)
			} else {
				m["vrange"] = "/"
			}
		}
	} else {
		m["vrange"] = "/"
	}

	if rule.reg != nil {
		m["regexp"] = fmt.Sprintf(
			`<code class="regexp" descp="%s">%s</code>`,
			html.EscapeString(fmt.Sprintf("`%s`", rule.reg.String())),
			html.EscapeString(rule.regName),
		)
	} else {
		m["regexp"] = "/"
	}

	if rule.defaultFunc != nil {
		v, _ := jsonx.Marshal(rule.defaultFunc())
		m["default"] = html.EscapeString(fmt.Sprintf("%v", string(v)))
	} else {
		m["default"] = "/"
	}

	if rule.fnNames != "" {
		m["function"] = fmt.Sprintf(
			`<code class="function">%s</code>`,
			html.EscapeString(rule.fnNames),
		)
	} else {
		m["function"] = "/"
	}

	return ruleFmt.Render(m)
}

func (rule *_Rule) toBytes(v []byte) ([]byte, bool) {
	if rule.checkFieldBytesSize {
		i := len(v)
		if rule.minFieldBytesSize != nil && i < *rule.minFieldBytesSize {
			return nil, false
		}
		if rule.maxFieldBytesSize != nil && i > *rule.maxFieldBytesSize {
			return nil, false
		}
	}

	if !rule.notTrimSpace {
		v = bytes.TrimSpace(v)
	}

	if rule.reg != nil && !rule.reg.Match(v) {
		return nil, false
	}

	var ok bool
	for _, f := range rule.fns {
		v, ok = f(v)
		if !ok {
			return nil, false
		}
	}
	return v, true
}

func (rule *_Rule) toString(v []byte) (string, bool) {
	v, ok := rule.toBytes(v)
	if !ok {
		return "", false
	}
	if rule.notEscapeHtml {
		return utils.S(v), true
	}
	return utils.S(htmlEscape(v)), true
}

func (rule *_Rule) toInt(v []byte) (int64, bool) {
	i, e := strconv.ParseInt(utils.S(v), 10, 64)
	if e != nil {
		return 0, false
	}

	if rule.checkNumRange {
		if rule.minIntVal != nil && i < *rule.minIntVal {
			return 0, false
		}
		if rule.maxIntVal != nil && i > *rule.maxIntVal {
			return 0, false
		}
	}
	return i, true
}

func (rule *_Rule) toUint(v []byte) (uint64, bool) {
	i, e := strconv.ParseUint(utils.S(v), 10, 64)
	if e != nil {
		return 0, false
	}

	if rule.checkNumRange {
		if rule.minUintVal != nil && i < *rule.minUintVal {
			return 0, false
		}
		if rule.maxUintVal != nil && i > *rule.maxUintVal {
			return 0, false
		}
	}
	return i, true
}

func (rule *_Rule) toFloat(v []byte) (float64, bool) {
	i, e := strconv.ParseFloat(utils.S(v), 64)
	if e != nil {
		return 0, false
	}

	if rule.checkNumRange {
		if rule.minDoubleVal != nil && i < *rule.minDoubleVal {
			return 0, false
		}
		if rule.maxDoubleVal != nil && i > *rule.maxDoubleVal {
			return 0, false
		}
	}
	return i, true
}

var ParseBool func(v []byte) (bool, error)

func init() {
	ParseBool = func(v []byte) (bool, error) {
		return strconv.ParseBool(utils.S(v))
	}
}

func (rule *_Rule) toBool(v []byte) (bool, bool) {
	b, e := ParseBool(v)
	if e != nil {
		return false, false
	}
	return b, true
}

type Rules []*_Rule

func (rules Rules) String() string {
	buf := strings.Builder{}
	buf.WriteString(MarkdownTableHeader)
	for _, r := range rules {
		buf.WriteString(r.String())
		buf.WriteByte('\n')
	}
	return buf.String()
}
