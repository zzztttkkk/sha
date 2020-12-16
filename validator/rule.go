package validator

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/sha/internal"
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
	_WhereParams
	_WhereQuery
	_WhereBody
	_WhereCookie
	_WhereHeader
)

type Rule struct {
	fieldIndex          []int
	fieldType           reflect.Type
	indirectCustomField bool

	formName       string
	pathParamsName string
	rtype          _RuleType

	isRequired  bool
	description string

	fVR   bool // int value range flag
	minIV *int64
	maxIV *int64
	minUV *uint64
	maxUV *uint64
	minDV *float64
	maxDV *float64

	isSlice bool
	fSSR    bool // slice size range flag
	minSSV  *int
	maxSSV  *int

	notEscapeHtml bool
	notTrimSpace  bool
	fLR           bool // bytes size range flag
	minLV         *int
	maxLV         *int

	defaultVal interface{}

	reg     *regexp.Regexp
	regName string

	fn     func([]byte) ([]byte, bool)
	fnName string
}

var MarkdownTableHeader = "\n|name|type|required|string length range|int value range|list size range|default|regexp|function|description|\n"

func init() {
	MarkdownTableHeader += strings.Repeat("|:---:", 10)
	MarkdownTableHeader += "|\n"
}

var ruleFmt = internal.NewNamedFmt(
	"|${name}|${type}|${required}|${lrange}|${vrange}|${srange}|${default}|${regexp}|${function}|${description}|",
)

// markdown table row
func (rule *Rule) String() string {
	typeString := typeNames[rule.rtype]
	if rule.rtype == _CustomType {
		if rule.indirectCustomField {
			typeString = reflect.New(rule.fieldType).Type().Elem().Name()
		} else {
			typeString = rule.fieldType.Name()
		}
	}
	m := internal.M{
		"type":     typeString,
		"required": fmt.Sprintf("%v", rule.isRequired),
	}

	if len(rule.description) > 0 {
		m["description"] = rule.description
	} else {
		m["description"] = "/"
	}

	m["name"] = string(rule.formName)
	if len(rule.pathParamsName) > 0 {
		m["name"] = fmt.Sprintf("PathParam: %s", rule.pathParamsName)
	}

	if rule.fLR {
		if rule.minLV != nil && rule.maxLV != nil {
			m["lrange"] = fmt.Sprintf("%d-%d", *rule.minLV, *rule.maxLV)
		} else if rule.minLV != nil {
			m["lrange"] = fmt.Sprintf("%d-", *rule.minLV)
		} else if rule.maxLV != nil {
			m["lrange"] = fmt.Sprintf("-%d", *rule.maxLV)
		} else {
			m["lrange"] = "/"
		}
	} else {
		m["lrange"] = "/"
	}

	if rule.fSSR {
		if rule.minSSV != nil && rule.maxSSV != nil {
			m["srange"] = fmt.Sprintf("%d-%d", *rule.minSSV, *rule.maxSSV)
		} else if rule.minSSV != nil {
			m["srange"] = fmt.Sprintf("%d-", *rule.minSSV)
		} else if rule.maxSSV != nil {
			m["srange"] = fmt.Sprintf("-%d", *rule.maxSSV)
		} else {
			m["srange"] = "/"
		}
	} else {
		m["srange"] = "/"
	}

	if rule.fVR {
		switch rule.rtype {
		case _Int64, _IntSlice:
			if rule.minIV != nil && rule.maxIV != nil {
				m["vrange"] = fmt.Sprintf("%d-%d", *rule.minIV, *rule.maxIV)
			} else if rule.minIV != nil {
				m["vrange"] = fmt.Sprintf("%d-", *rule.minIV)
			} else if rule.maxIV != nil {
				m["vrange"] = fmt.Sprintf("-%d", *rule.maxIV)
			} else {
				m["vrange"] = "/"
			}
		case _Uint64, _UintSlice:
			if rule.minUV != nil && rule.maxUV != nil {
				m["vrange"] = fmt.Sprintf("%d-%d", *rule.minUV, *rule.maxUV)
			} else if rule.minUV != nil {
				m["vrange"] = fmt.Sprintf("%d-", *rule.minUV)
			} else if rule.maxUV != nil {
				m["vrange"] = fmt.Sprintf("-%d", *rule.maxUV)
			} else {
				m["vrange"] = "/"
			}
		case _Float64, _FloatSlice:
			if rule.minDV != nil && rule.maxDV != nil {
				m["vrange"] = fmt.Sprintf("%f-%f", *rule.minDV, *rule.maxDV)
			} else if rule.minDV != nil {
				m["vrange"] = fmt.Sprintf("%f-", *rule.minDV)
			} else if rule.maxDV != nil {
				m["vrange"] = fmt.Sprintf("-%f", *rule.maxDV)
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

	if rule.defaultVal != nil {
		m["default"] = html.EscapeString(fmt.Sprintf("%v", rule.defaultVal))
	} else {
		m["default"] = "/"
	}

	if rule.fnName != "" {
		m["function"] = fmt.Sprintf(
			`<code class="function" descp="%s">%s</code>`,
			html.EscapeString(bytesFilterDescriptionMap[rule.fnName]),
			html.EscapeString(rule.fnName),
		)
	} else {
		m["function"] = "/"
	}

	return ruleFmt.Render(m)
}

func (rule *Rule) toBytes(v []byte) ([]byte, bool) {
	if !rule.notTrimSpace {
		v = bytes.TrimSpace(v)
	}

	if rule.fLR {
		l := len(v)
		if rule.minLV != nil && l < *rule.minLV {
			return nil, false
		}
		if rule.maxLV != nil && l > *rule.maxLV {
			return nil, false
		}
	}

	if rule.reg != nil {
		if !rule.reg.Match(v) {
			return nil, false
		}
	}

	if rule.fn != nil {
		var ok bool
		v, ok = rule.fn(v)
		if !ok {
			return nil, false
		}
	}
	return v, true
}

func (rule *Rule) toString(v []byte) (string, bool) {
	v, ok := rule.toBytes(v)
	if !ok {
		return "", false
	}
	if rule.notEscapeHtml {
		return internal.S(v), true
	}
	return internal.S(htmlEscape(v)), true
}

func (rule *Rule) toInt(v []byte) (int64, bool) {
	var ok bool
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}

	i, e := strconv.ParseInt(internal.S(v), 10, 64)
	if e != nil {
		return 0, false
	}

	if rule.fVR {
		if rule.minIV != nil && i < *rule.minIV {
			return 0, false
		}
		if rule.maxIV != nil && i > *rule.maxIV {
			return 0, false
		}
	}
	return i, true
}

func (rule *Rule) toUint(v []byte) (uint64, bool) {
	var ok bool
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}

	i, e := strconv.ParseUint(internal.S(v), 10, 64)
	if e != nil {
		return 0, false
	}

	if rule.fVR {
		if rule.minUV != nil && i < *rule.minUV {
			return 0, false
		}
		if rule.maxUV != nil && i > *rule.maxUV {
			return 0, false
		}
	}
	return i, true
}

func (rule *Rule) toFloat(v []byte) (float64, bool) {
	var ok bool
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}

	i, e := strconv.ParseFloat(internal.S(v), 64)
	if e != nil {
		return 0, false
	}

	if rule.fVR {
		if rule.minDV != nil && i < *rule.minDV {
			return 0, false
		}
		if rule.maxDV != nil && i > *rule.maxDV {
			return 0, false
		}
	}
	return i, true
}

var ParseBool func(v []byte) (bool, error)

func (rule *Rule) toBool(v []byte) (bool, bool) {
	var ok bool
	v, ok = rule.toBytes(v)
	if !ok {
		return false, false
	}

	var b bool
	var e error
	if ParseBool == nil {
		b, e = strconv.ParseBool(internal.S(v))
	} else {
		b, e = ParseBool(v)
	}
	if e != nil {
		return false, false
	}
	return b, true
}

type Rules []*Rule

func (rules Rules) String() string {
	buf := strings.Builder{}

	buf.WriteString("#### fields\n")
	buf.WriteString(MarkdownTableHeader)
	for _, r := range rules {
		buf.WriteString(r.String())
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (rules Rules) Document() string {
	return rules.String()
}
