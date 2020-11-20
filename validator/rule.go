package validator

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"html"
	"regexp"
	"strconv"
	"strings"
)

type _RuleType int

const (
	Bool = _RuleType(iota)
	Int64
	Uint64
	Float64
	Bytes
	String

	BoolSlice
	IntSlice
	UintSlice
	FloatSlice
	StringSlice
	BytesSlice
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
}

type Rule struct {
	fieldIndex     []int
	formName       []byte
	pathParamsName []byte
	rtype          _RuleType

	isRequired bool

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

	notTrimSpace bool
	fLR          bool // bytes size range flag
	minLV        *int
	maxLV        *int

	defaultValPtr *[]byte

	reg     *regexp.Regexp
	regName string

	fn     func([]byte) ([]byte, bool)
	fnName string
}

var MarkdownTableHeader = "\n|name|type|required|length range|value range|size range|default|regexp|function|\n"

func init() {
	MarkdownTableHeader += strings.Repeat("|:---:", 9)
	MarkdownTableHeader += "|\n"
}

var ruleFmt = internal.NewNamedFmt(
	"|${name}|${type}|${required}|${lrange}|${vrange}|${srange}|${default}|${regexp}|${function}|",
)

// markdown table row
func (rule *Rule) String() string {
	m := internal.M{
		"type":     typeNames[rule.rtype],
		"required": fmt.Sprintf("%v", rule.isRequired),
	}

	m["name"] = string(rule.formName)
	if len(rule.pathParamsName) > 0 {
		m["name"] = fmt.Sprintf("PathParam: %s", rule.pathParamsName)
	}
	if rule.fSSR {
		m["type"] = fmt.Sprintf("%s; multi values", m["type"])
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
		case Int64, IntSlice:
			if rule.minIV != nil && rule.maxIV != nil {
				m["vrange"] = fmt.Sprintf("%d-%d", *rule.minIV, *rule.maxIV)
			} else if rule.minIV != nil {
				m["vrange"] = fmt.Sprintf("%d-", *rule.minIV)
			} else if rule.maxIV != nil {
				m["vrange"] = fmt.Sprintf("-%d", *rule.maxIV)
			} else {
				m["vrange"] = "/"
			}
		case Uint64, UintSlice:
			if rule.minUV != nil && rule.maxUV != nil {
				m["vrange"] = fmt.Sprintf("%d-%d", *rule.minUV, *rule.maxUV)
			} else if rule.minUV != nil {
				m["vrange"] = fmt.Sprintf("%d-", *rule.minUV)
			} else if rule.maxUV != nil {
				m["vrange"] = fmt.Sprintf("-%d", *rule.maxUV)
			} else {
				m["vrange"] = "/"
			}
		case Float64, FloatSlice:
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

	if rule.defaultValPtr != nil && len(*rule.defaultValPtr) > 0 {
		m["default"] = html.EscapeString(string(*rule.defaultValPtr))
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
	return internal.S(v), true
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

	descp := descriptionMap[fmt.Sprintf("%p", rules)]
	if len(descp) > 0 {
		buf.WriteString(
			fmt.Sprintf("#### description\n\n%s\n\n", descp),
		)
	}

	buf.WriteString("#### fields\n\n")
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
