package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/utils"
	"html"
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
	_JsonObject
	_BoolSlice
	_IntSlice
	_UintSlice
	_StringSlice
)

var typeNames = []string{
	"Bool",
	"Int",
	"Uint",
	"String",
	"String",
	"JsonObject",
	"BoolArray",
	"IntArray",
	"UintArray",
	"StringArray",
}

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

	reg     *regexp.Regexp
	regName string
	fn      func([]byte) ([]byte, bool)
	fnName  string
}

var TrueBytes = []byte("true")

func (rule *_RuleT) toBytes(v []byte) (val []byte, ok bool) {
	if rule.lrange {
		l := len(v)
		if rule.minL > 0 && l < rule.minL {
			return nil, false
		}
		if rule.maxL > 0 && l > rule.maxL {
			return nil, false
		}
	}

	if rule.reg != nil {
		if !rule.reg.Match(v) {
			return nil, false
		}
	}

	if rule.fn != nil {
		v, ok = rule.fn(v)
		if !ok {
			return nil, false
		}
	}

	return v, true
}

func (rule *_RuleT) toBool(v []byte) (r bool, ok bool) {
	v, ok = rule.toBytes(v)
	if !ok {
		return false, false
	}
	return bytes.Equal(v, TrueBytes), true
}

func (rule *_RuleT) toI64(v []byte) (num int64, ok bool) {
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}

	rv, err := strconv.ParseInt(utils.B2s(v), 10, 64)
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

func (rule *_RuleT) toUI64(v []byte) (num uint64, ok bool) {
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}

	rv, err := strconv.ParseUint(utils.B2s(v), 10, 64)
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

func (rule *_RuleT) toJsonObj(v []byte) (map[string]interface{}, bool) {
	if rule.lrange {
		l := len(v)
		if rule.minL > 0 && l < rule.minL {
			return nil, false
		}
		if rule.maxL > 0 && l > rule.maxL {
			return nil, false
		}
	}

	m := map[string]interface{}{}
	err := json.Unmarshal(v, &m)
	if err != nil {
		return nil, false
	}
	return m, true
}

func mapMultiForm(ctx *fasthttp.RequestCtx, name string, fn func([]byte) bool) bool {
	for _, v := range ctx.QueryArgs().PeekMulti(name) {
		if !fn(v) {
			return false
		}
	}

	for _, v := range ctx.PostArgs().PeekMulti(name) {
		if !fn(v) {
			return false
		}
	}

	mf, _ := ctx.MultipartForm()
	if mf != nil {
		for _, v := range mf.Value[name] {
			if !fn(utils.S2b(v)) {
				return false
			}
		}
	}
	return true
}

func (rule *_RuleT) toBoolSlice(ctx *fasthttp.RequestCtx) ([]bool, bool) {
	var lst []bool
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toBool(i)
			if !ok {
				return false
			}
			lst = append(lst, v)
			return true
		},
	)
	if ok {
		return lst, true
	}
	return nil, false
}

func (rule *_RuleT) toIntSlice(ctx *fasthttp.RequestCtx) ([]int64, bool) {
	var lst []int64
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toI64(i)
			if !ok {
				return false
			}
			lst = append(lst, v)
			return true
		},
	)
	if ok {
		return lst, true
	}
	return nil, false
}

func (rule *_RuleT) toUintSlice(ctx *fasthttp.RequestCtx) ([]uint64, bool) {
	var lst []uint64
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toUI64(i)
			if !ok {
				return false
			}
			lst = append(lst, v)
			return true
		},
	)
	if ok {
		return lst, true
	}
	return nil, false
}

func (rule *_RuleT) toStrSlice(ctx *fasthttp.RequestCtx) ([]string, bool) {
	var lst []string
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toBytes(i)
			if !ok {
				return false
			}
			lst = append(lst, utils.B2s(v))
			return true
		},
	)
	if ok {
		return lst, true
	}
	return nil, false
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
		m["regexp"] = fmt.Sprintf(
			`<code class="regexp" descp="%s">%s</code>`,
			html.EscapeString(fmt.Sprintf("`%s`", rule.reg.String())),
			html.EscapeString(rule.regName),
		)
	} else {
		m["regexp"] = "/"
	}

	if len(rule.defaultV) > 0 {
		m["default"] = html.EscapeString(string(rule.defaultV))
	} else {
		m["default"] = "/"
	}

	if rule.fnName != "" {
		m["function"] = fmt.Sprintf(
			`<code class="function" descp="%s">%s</cpde>`,
			html.EscapeString(funcDescp[rule.fnName]),
			html.EscapeString(rule.fnName),
		)
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

// markdown table
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
