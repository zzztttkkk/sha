package validator

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/snow/internal"
	"github.com/zzztttkkk/snow/utils"
	"regexp"
	"strconv"
	"strings"
)

const (
	_Bool = iota
	_Int64
	_Uint64
	_Bytes
	_String
)

var typeNames = []string{"bool", "int", "uint", "string", "string"}

type _RuleT struct {
	form     string
	field    string
	t        int
	required bool

	vrange bool
	minV   int64
	maxV   int64
	minUV  uint64
	maxUV  uint64

	lrange bool
	minL   int
	maxL   int

	defaultV []byte

	reg    *regexp.Regexp
	fn     func([]byte) ([]byte, bool)
	fnName string
}

var trueBytes = []byte("true")

func (rule *_RuleT) toBool(v []byte) bool {
	return bytes.Equal(v, trueBytes)
}

func (rule *_RuleT) toI64(v []byte) (int64, bool) {
	rv, err := strconv.ParseInt(internal.B2s(v), 10, 64)
	if err != nil {
		return 0, false
	}

	if rule.vrange {
		if rv < rule.minV {
			return 0, false
		}

		if rv > rule.maxV {
			return 0, false
		}
	}

	return rv, true
}

func (rule *_RuleT) toUI64(v []byte) (uint64, bool) {
	rv, err := strconv.ParseUint(internal.B2s(v), 10, 64)
	if err != nil {
		return 0, false
	}

	if rule.vrange {
		if rv < rule.minUV {
			return 0, false
		}

		if rv > rule.maxUV {
			return 0, false
		}
	}

	return rv, true
}

func (rule *_RuleT) toBytes(v []byte) ([]byte, bool) {
	if rule.lrange {
		l := len(v)
		if rule.minL > 0 && l < rule.minL {
			return nil, false
		}
		if rule.maxL > 0 && l > rule.maxL {
			return nil, false
		}
	}

	if rule.fn != nil {
		return rule.fn(v)
	}

	if rule.reg != nil {
		return v, rule.reg.Match(v)
	}

	return v, true
}

var ruleFmt = utils.NewNamedFmtCached("|{name}|{type}|{required}|{lrange}|{vrange}|{default}|{regexp}|{function}|")

func (rule *_RuleT) String() string {
	m := utils.M{
		"name":     rule.form,
		"type":     typeNames[rule.t],
		"required": rule.required,
	}

	if rule.vrange {
		if rule.t == _Int64 {
			m["vrange"] = fmt.Sprintf("%d-%d", rule.minV, rule.maxV)
		} else {
			m["vrange"] = fmt.Sprintf("%d-%d", rule.minUV, rule.maxUV)
		}
	} else {
		m["vrange"] = "/"
	}

	if rule.lrange {
		m["lrange"] = fmt.Sprintf("%d-%d", rule.minL, rule.maxL)
	} else {
		m["lrange"] = "/"
	}

	if rule.reg != nil {
		m["regexp"] = fmt.Sprintf("`%s`", rule.reg.String())
	} else {
		m["regexp"] = "/"
	}

	if len(rule.defaultV) > 0 {
		m["default"] = string(rule.defaultV)
	} else {
		m["default"] = "/"
	}

	if rule.fnName != "" {
		m["function"] = rule.fnName
	} else {
		m["function"] = "/"
	}
	return ruleFmt.Render(m)
}

type _RuleSliceT []*_RuleT

func (a _RuleSliceT) Len() int      { return len(a) }
func (a _RuleSliceT) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a _RuleSliceT) Less(i, j int) bool {
	l, r := a[i], a[j]
	if l.required != r.required {
		if l.required {
			return true
		}
	}
	return a[i].form < a[j].form
}

func (a _RuleSliceT) String() string {
	buf := strings.Builder{}
	buf.WriteString("|name|type|required|lrange|vrange|default|regexp|function|\n")
	buf.WriteString("|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|\n")
	for _, r := range a {
		buf.WriteString(r.String())
		buf.WriteByte('\n')
	}
	return buf.String()
}
