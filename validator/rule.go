package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	minIV         *int64
	maxIV         *int64
	minUV         *uint64
	maxUV         *uint64
	minDV         *float64
	maxDV         *float64

	isSlice       bool
	checkListSize bool // slice size range flag
	minSSV        *int
	maxSSV        *int

	notEscapeHtml bool
	notTrimSpace  bool

	defaultFunc func() interface{}

	reg     *regexp.Regexp
	regName string

	fns     []func([]byte) ([]byte, bool)
	fnNames string
}

var MarkdownTableHeader = "\n|name|type|required|int value range|list size range|default|regexp|function|description|\n"

func init() {
	MarkdownTableHeader += strings.Repeat("|:---:", 9)
	MarkdownTableHeader += "|\n"
}

var ruleFmt = utils.NewNamedFmt(
	"|${name}|${type}|${required}|${vrange}|${srange}|${default}|${regexp}|${function}|${description}|",
)

// markdown table row
func (rule *_Rule) String() string {
	typeString := typeNames[rule.rtype]
	if rule.rtype == _CustomType {
		if rule.isPtr {
			typeString = reflect.New(rule.fieldType).Type().Elem().Name()
		} else {
			typeString = rule.fieldType.Name()
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
		m["name"] = fmt.Sprintf("Form<%s>", rule.formName)
	case _WhereCookie:
		m["name"] = fmt.Sprintf("Cookie<%s>", rule.formName)
	case _WhereHeader:
		m["name"] = fmt.Sprintf("Header<%s>", rule.formName)
	case _WhereBody:
		m["name"] = fmt.Sprintf("Body<%s>", rule.formName)
	case _WhereQuery:
		m["name"] = fmt.Sprintf("Query<%s>", rule.formName)
	case _WhereURLParams:
		m["name"] = fmt.Sprintf("URLParams<%s>", rule.formName)
	}

	if rule.checkListSize {
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

	if rule.checkNumRange {
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

	if rule.defaultFunc != nil {
		v, _ := json.Marshal(rule.defaultFunc())
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
		if rule.minIV != nil && i < *rule.minIV {
			return 0, false
		}
		if rule.maxIV != nil && i > *rule.maxIV {
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
		if rule.minUV != nil && i < *rule.minUV {
			return 0, false
		}
		if rule.maxUV != nil && i > *rule.maxUV {
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

	buf.WriteString("#### fields\n")
	buf.WriteString(MarkdownTableHeader)
	for _, r := range rules {
		buf.WriteString(r.String())
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (rules Rules) Len() int { return len(rules) }

func (rules Rules) Less(i, j int) bool { return rules[i].where < rules[j].where }

func (rules Rules) Swap(i, j int) { rules[i], rules[j] = rules[j], rules[i] }
