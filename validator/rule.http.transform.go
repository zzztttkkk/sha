package validator

import (
	"reflect"
	"strconv"

	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/jsonx"
)

func (rule *_Rule) toBytes(v []byte) (val []byte, ok bool) {
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

func (rule *_Rule) toBool(v []byte) (r bool, ok bool) {
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

func (rule *_Rule) toInt(v []byte) (num int64, ok bool) {
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

func (rule *_Rule) toUint(v []byte) (num uint64, ok bool) {
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

func (rule *_Rule) toFloat(v []byte) (num float64, ok bool) {
	v, ok = rule.toBytes(v)
	if !ok {
		return 0, false
	}
	rv, err := strconv.ParseFloat(gotils.B2S(v), 10)
	if err != nil {
		return 0, false
	}
	if rule.vrange {
		if rule.minFVF && rv < rule.minFV {
			return 0, false
		}

		if rule.maxFVF && rv > rule.maxFV {
			return 0, false
		}
	}
	return rv, true
}

func (rule *_Rule) toJsonObj(v []byte) (map[string]interface{}, bool) {
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
	err := jsonx.Unmarshal(v, &m)
	if err != nil {
		return nil, false
	}
	return m, true
}

func (rule *_Rule) toJsonAry(v []byte) ([]interface{}, bool) {
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
	err := jsonx.Unmarshal(v, &s)
	if err != nil {
		return nil, false
	}
	return s, true
}

func (rule *_Rule) checkSizeRange(v *reflect.Value) bool {
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

func (rule *_Rule) toBoolSlice(ctx *fasthttp.RequestCtx) ([]bool, bool) {
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

func (rule *_Rule) toIntSlice(ctx *fasthttp.RequestCtx) ([]int64, bool) {
	var lst []int64
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toInt(i)
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

func (rule *_Rule) toUintSlice(ctx *fasthttp.RequestCtx) ([]uint64, bool) {
	var lst []uint64
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toUint(i)
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

func (rule *_Rule) toFloatSlice(ctx *fasthttp.RequestCtx) ([]float64, bool) {
	var lst []float64
	ok := mapMultiForm(
		ctx,
		rule.form,
		func(i []byte) bool {
			v, ok := rule.toFloat(i)
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

func (rule *_Rule) toStrSlice(ctx *fasthttp.RequestCtx) ([]string, bool) {
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
