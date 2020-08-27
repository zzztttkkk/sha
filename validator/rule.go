package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/utils"
	"html"
	"reflect"
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
	_JsonArray

	_BoolSlice
	_IntSlice
	_UintSlice
	_StringSlice
	_BytesSlice
)

var typeNames = []string{
	"Bool",
	"Int",
	"Uint",
	"String",
	"String",

	"JsonObject",
	"JsonArray",

	"BoolArray",
	"IntArray",
	"UintArray",
	"StringArray",
	"StringArray",
}

type _RuleT struct {
	form     string
	field    string
	t        int
	required bool

	vrange bool // int value value range
	minVF  bool
	minV   int64
	maxVF  bool
	maxV   int64
	minUVF bool
	minUV  uint64
	maxUVF bool
	maxUV  uint64

	lrange bool // bytes value length range
	minLF  bool
	minL   int64
	maxLF  bool
	maxL   int64

	srange bool // slice value size range
	minSF  bool
	minS   int64
	maxSF  bool
	maxS   int64

	defaultV []byte

	reg     *regexp.Regexp
	regName string

	fn     func([]byte) ([]byte, bool)
	fnName string

	isSlice bool
}

func (rule *_RuleT) toBytes(v []byte) (val []byte, ok bool) {
	if rule.lrange {
		l := int64(len(v))

		if rule.minLF && l < rule.minL {
			return nil, false
		}

		if rule.maxLF && l > rule.maxL {
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

	_v, err := strconv.ParseBool(gotils.B2S(v))
	if err != nil {
		return false, false
	}
	return _v, true
}

func (rule *_RuleT) toI64(v []byte) (num int64, ok bool) {
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}

	rv, err := strconv.ParseInt(gotils.B2S(v), 10, 64)
	if err != nil {
		return 0, false
	}

	if rule.vrange {
		if rule.minVF && rv < rule.minV {
			return 0, false
		}
		if rule.maxVF && rv > rule.maxV {
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

	rv, err := strconv.ParseUint(gotils.B2S(v), 10, 64)
	if err != nil {
		return 0, false
	}

	if rule.vrange {
		if rule.minUVF && rv < rule.minUV {
			return 0, false
		}

		if rule.maxUVF && rv > rule.maxUV {
			return 0, false
		}
	}

	return rv, true
}

func (rule *_RuleT) toJsonObj(v []byte) (map[string]interface{}, bool) {
	if rule.lrange {
		l := int64(len(v))
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

func (rule *_RuleT) toJsonAry(v []byte) ([]interface{}, bool) {
	if rule.lrange {
		l := int64(len(v))
		if rule.minL > 0 && l < rule.minL {
			return nil, false
		}
		if rule.maxL > 0 && l > rule.maxL {
			return nil, false
		}
	}

	var s []interface{}
	err := json.Unmarshal(v, &s)
	if err != nil {
		return nil, false
	}
	return s, true
}

func (rule *_RuleT) checkSize(v *reflect.Value) bool {
	if !rule.srange {
		return true
	}
	_l := int64(v.Len())

	if rule.minSF && _l < rule.minS {
		return false
	}
	if rule.maxSF && _l > rule.maxS {
		return false
	}
	return true
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
			if !fn(gotils.S2B(v)) {
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

var joinSep = []byte(",")

func (rule *_RuleT) toJoinedBoolSlice(ctx *fasthttp.RequestCtx) (lst []bool, ok bool) {
	formV := ctx.FormValue(rule.form)
	if len(formV) < 1 {
		return nil, false
	}
	for _, b := range bytes.Split(formV, joinSep) {
		_v, _ok := rule.toBool(bytes.TrimSpace(b))
		if !_ok {
			return nil, false
		}
		lst = append(lst, _v)
	}
	return lst, true
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

func (rule *_RuleT) toJoinedIntSlice(ctx *fasthttp.RequestCtx) (lst []int64, ok bool) {
	formV := ctx.FormValue(rule.form)
	if len(formV) < 1 {
		return nil, false
	}
	for _, b := range bytes.Split(formV, joinSep) {
		_v, _ok := rule.toI64(bytes.TrimSpace(b))
		if !_ok {
			return nil, false
		}
		lst = append(lst, _v)
	}
	return lst, true
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

func (rule *_RuleT) toJoinedUintSlice(ctx *fasthttp.RequestCtx) (lst []uint64, ok bool) {
	formV := ctx.FormValue(rule.form)
	if len(formV) < 1 {
		return nil, false
	}
	for _, b := range bytes.Split(formV, joinSep) {
		_v, _ok := rule.toUI64(bytes.TrimSpace(b))
		if !_ok {
			return nil, false
		}
		lst = append(lst, _v)
	}
	return lst, true
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
			lst = append(lst, gotils.B2S(v))
			return true
		},
	)
	if ok {
		return lst, true
	}
	return nil, false
}

func (rule *_RuleT) ToGraphqlArgument() *graphql.ArgumentConfig {
	ac := &graphql.ArgumentConfig{}

	switch rule.t {
	case _Bool:
		ac.Type = graphql.Boolean
	case _Int64, _Uint64:
		ac.Type = graphql.Int
	case _Bytes, _String:
		ac.Type = graphql.String
	}
	return ac
}

var ruleFmt = utils.NewNamedFmt("|{name}|{type}|{required}|{lrange}|{vrange}|{srange}|{default}|{regexp}|{function}|")

func (rule *_RuleT) String() string {
	m := utils.M{
		"name":     rule.form,
		"type":     typeNames[rule.t],
		"required": rule.required,
	}

	if rule.vrange {
		if rule.t == _Int64 {
			if rule.minVF && rule.maxVF {
				m["vrange"] = fmt.Sprintf("%d-%d", rule.minV, rule.maxV)
			} else if rule.minVF {
				m["vrange"] = fmt.Sprintf("%d-", rule.minV)
			} else if rule.maxVF {
				m["vrange"] = fmt.Sprintf("-%d", rule.maxV)
			} else {
				m["vrange"] = "/"
			}
		} else {
			if rule.minUVF && rule.maxUVF {
				m["vrange"] = fmt.Sprintf("%d-%d", rule.minUV, rule.maxUV)
			} else if rule.minUVF {
				m["vrange"] = fmt.Sprintf("%d-", rule.minUV)
			} else if rule.maxUVF {
				m["vrange"] = fmt.Sprintf("-%d", rule.maxUV)
			} else {
				m["vrange"] = "/"
			}
		}
	} else {
		m["vrange"] = "/"
	}

	if rule.lrange {
		if rule.minLF && rule.maxLF {
			m["lrange"] = fmt.Sprintf("%d-%d", rule.minL, rule.maxL)
		} else if rule.minLF {
			m["lrange"] = fmt.Sprintf("%d-", rule.minL)
		} else if rule.maxUVF {
			m["lrange"] = fmt.Sprintf("-%d", rule.maxL)
		} else {
			m["lrange"] = "/"
		}
	} else {
		m["lrange"] = "/"
	}

	if rule.srange {
		if rule.minSF && rule.maxSF {
			m["srange"] = fmt.Sprintf("%d-%d", rule.minS, rule.maxS)
		} else if rule.minSF {
			m["srange"] = fmt.Sprintf("%d-", rule.minS)
		} else if rule.maxSF {
			m["srange"] = fmt.Sprintf("-%d", rule.maxS)
		} else {
			m["srange"] = "/"
		}
	} else {
		m["srange"] = "/"
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
	buf.WriteString("|name|type|required|length range|value range|size range|default|regexp|function|\n")
	buf.WriteString("|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|\n")
	for _, r := range a {
		buf.WriteString(r.String())
		buf.WriteByte('\n')
	}
	return buf.String()
}

type Rules struct {
	lst    _RuleSliceT
	isJson bool
	doc    *Doc
	raw    reflect.Type
}

func (rs *Rules) NewDoc(descp string) *Doc {
	return &Doc{
		descp:  descp,
		fields: rs.lst.String(),
	}
}

func (rs *Rules) RawType() reflect.Type {
	return rs.raw
}

type Doc struct {
	descp  string
	fields string
}

func (d *Doc) DocDescp() string { return d.descp }

func (d *Doc) ApiFields() string {
	return d.fields
}
