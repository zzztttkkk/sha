package validator

import (
	"github.com/graphql-go/graphql"
	"github.com/savsgio/gotils"
	"github.com/zzztttkkk/suna/jsonx"
	"math"
	"reflect"
)

func (rule *_Rule) toGraphqlArgument() *graphql.ArgumentConfig {
	cfg := &graphql.ArgumentConfig{Description: rule.info}
	switch rule.t {
	case _Bool:
		cfg.Type = graphql.Boolean
		if len(rule.defaultV) > 0 {
			v, _ := rule.toBool(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _Int64:
		cfg.Type = graphql.Int
		if len(rule.defaultV) > 0 {
			v, _ := rule.toI64(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _Uint64:
		cfg.Type = graphql.Int
		if len(rule.defaultV) > 0 {
			v, _ := rule.toUI64(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _String, _Bytes:
		cfg.Type = graphql.String
		if len(rule.defaultV) > 0 {
			v, _ := rule.toBytes(rule.defaultV)
			cfg.DefaultValue = string(v)
		}
	case _JsonArray:
		cfg.Type = jsonx.GraphqlArray
	case _JsonObject:
		cfg.Type = jsonx.GraphqlObject
	case _BoolSlice:
		cfg.Type = graphql.NewList(graphql.Boolean)
	case _IntSlice, _UintSlice:
		cfg.Type = graphql.NewList(graphql.Int)
	case _StringSlice, _BytesSlice:
		cfg.Type = graphql.NewList(graphql.String)
	}
	return cfg
}

func (rs *Rules) ArgumentMap() graphql.FieldConfigArgument {
	m := graphql.FieldConfigArgument{}
	for _, r := range rs.lst {
		m[r.form] = r.toGraphqlArgument()
	}
	return m
}

func valueToInt64(v interface{}) (int64, bool) {
	switch rv := v.(type) {
	case uint:
		if rv > math.MaxInt64 {
			break
		}
		return int64(rv), true
	case uint8:
		return int64(rv), true
	case uint16:
		return int64(rv), true
	case uint32:
		return int64(rv), true
	case uint64:
		if rv > math.MaxInt64 {
			break
		}
		return int64(rv), true
	case int:
		return int64(rv), true
	case int8:
		return int64(rv), true
	case int16:
		return int64(rv), true
	case int32:
		return int64(rv), true
	case int64:
		return int64(rv), true
	}
	return 0, false
}

func valueToUint64(v interface{}) (uint64, bool) {
	switch rv := v.(type) {
	case uint:
		return uint64(rv), true
	case uint8:
		return uint64(rv), true
	case uint16:
		return uint64(rv), true
	case uint32:
		return uint64(rv), true
	case uint64:
		return uint64(rv), true
	case int:
		if rv < 0 {
			return 0, false
		}
		return uint64(rv), true
	case int8:
		if rv < 0 {
			return 0, false
		}
		return uint64(rv), true
	case int16:
		if rv < 0 {
			return 0, false
		}
		return uint64(rv), true
	case int32:
		if rv < 0 {
			return 0, false
		}
		return uint64(rv), true
	case int64:
		if rv < 0 {
			return 0, false
		}
		return uint64(rv), true
	}
	return 0, false
}

func valueToBool(val interface{}) (bool, bool) {
	switch rv := val.(type) {
	case bool:
		return rv, true
	}
	return false, false
}

func valueToString(val interface{}) (string, bool) {
	switch rv := val.(type) {
	case string:
		return rv, true
	case []byte:
		return gotils.B2S(rv), true
	}
	return "", false
}

func valueToBytes(val interface{}) ([]byte, bool) {
	switch rv := val.(type) {
	case string:
		return gotils.S2B(rv), true
	case []byte:
		return rv, true
	}
	return nil, false
}

var i4st = reflect.TypeOf([]int64{})

func copyIntSlice(val interface{}) (reflect.Value, bool) {
	switch rv := val.(type) {
	case []int64:
		return reflect.ValueOf(rv), true
	case []int:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []int8:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []int16:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []int32:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []uint64:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n > math.MaxInt64 {
				return reflect.Value{}, false
			}
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []uint:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n > math.MaxInt64 {
				return reflect.Value{}, false
			}
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []uint16:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []uint32:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	}
	return reflect.Value{}, false
}

var u4st = reflect.TypeOf([]uint64{})

func copyUintSlice(val interface{}) (reflect.Value, bool) {
	switch rv := val.(type) {
	case []int64:
		v := reflect.MakeSlice(u4st, len(rv), cap(rv))
		for i, n := range rv {
			if n < 0 {
				return reflect.Value{}, false
			}
			v.Index(i).SetUint(uint64(n))
		}
		return v, true
	case []int:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n < 0 {
				return reflect.Value{}, false
			}
			v.Index(i).SetUint(uint64(n))
		}
		return v, true
	case []int8:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n < 0 {
				return reflect.Value{}, false
			}
			v.Index(i).SetUint(uint64(n))
		}
		return v, true
	case []int16:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n < 0 {
				return reflect.Value{}, false
			}
			v.Index(i).SetUint(uint64(n))
		}
		return v, true
	case []int32:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n < 0 {
				return reflect.Value{}, false
			}
			v.Index(i).SetUint(uint64(n))
		}
		return v, true
	case []uint64:
		return reflect.ValueOf(rv), true
	case []uint:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			if n > math.MaxInt64 {
				return reflect.Value{}, false
			}
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []uint16:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	case []uint32:
		v := reflect.MakeSlice(i4st, len(rv), cap(rv))
		for i, n := range rv {
			v.Index(i).SetInt(int64(n))
		}
		return v, true
	}
	return reflect.Value{}, false
}

var sst = reflect.TypeOf([]string{})

func copyStringSlice(val interface{}) (reflect.Value, bool) {
	switch rv := val.(type) {
	case []string:
		return reflect.ValueOf(rv), true
	case [][]byte:
		v := reflect.MakeSlice(sst, len(rv), cap(rv))
		for i, b := range rv {
			v.Index(i).SetString(string(b))
		}
		return v, true
	}
	return reflect.Value{}, false
}

var bsst = reflect.TypeOf([]string{})

func copyBytesSlice(val interface{}) (reflect.Value, bool) {
	switch rv := val.(type) {
	case [][]byte:
		return reflect.ValueOf(rv), true
	case []string:
		v := reflect.MakeSlice(bsst, len(rv), cap(rv))
		for i, b := range rv {
			v.Index(i).SetBytes([]byte(b))
		}
		return v, true
	}
	return reflect.Value{}, false
}

func (rs *Rules) BindValue(m map[string]interface{}) reflect.Value {
	ele := reflect.New(rs.raw).Elem()
	for _, rule := range rs.lst {
		field := ele.FieldByName(rule.field)
		value, ok := m[rule.form]
		if !ok {
			continue
		}

		switch rule.t {
		case _String:
			v, ok := valueToString(value)
			if ok {
				field.SetString(v)
			}
		case _Uint64:
			v, ok := valueToUint64(value)
			if ok {
				field.SetUint(v)
			}
		case _Int64:
			v, ok := valueToInt64(value)
			if ok {
				field.SetInt(v)
			}
		case _Bool:
			v, ok := valueToBool(value)
			if ok {
				field.SetBool(v)
			}
		case _Bytes:
			v, ok := valueToBytes(value)
			if ok {
				field.SetBytes(v)
			}

		case _IntSlice:
			v, ok := copyIntSlice(value)
			if ok {
				field.Set(v)
			}
		case _UintSlice:
			v, ok := copyUintSlice(value)
			if ok {
				field.Set(v)
			}
		case _BoolSlice:
			switch value.(type) {
			case []bool:
				field.Set(reflect.ValueOf(value))
			}
		case _StringSlice:
			v, ok := copyStringSlice(value)
			if ok {
				field.Set(v)
			}
		case _BytesSlice:
			v, ok := copyBytesSlice(value)
			if ok {
				field.Set(v)
			}
		}
	}
	return ele
}
