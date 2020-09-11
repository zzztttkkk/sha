package validator

import (
	"errors"
	"reflect"

	"github.com/savsgio/gotils"
)

func (rule *_Rule) checkVRangeByReflect(v interface{}) bool {
	if !rule.vrange {
		return true
	}

	switch rv := v.(type) {
	case uint64:
		if rule.minUVF && rv < rule.minUV {
			return false
		}
		if rule.maxUVF && rv > rule.maxUV {
			return false
		}
	case int64:
		if rule.minVF && rv < rule.minV {
			return false
		}
		if rule.maxVF && rv > rule.maxV {
			return false
		}
	case float64:
		if rule.minFVF && rv < rule.minFV {
			return false
		}

		if rule.maxFVF && rv > rule.maxFV {
			return false
		}
	}
	return true
}

var ErrGraphqlType = errors.New("suna.validator: unexpected type")

//revive:disable:cyclomatic
func (rs *Rules) ValidateAndBind(m map[string]interface{}) (*reflect.Value, error) {
	ele := reflect.New(rs.raw).Elem()
	for _, rule := range rs.lst {
		field := ele.FieldByName(rule.field)
		value, ok := m[rule.form]
		if !ok {
			if rule.required {
				return nil, rule.toFormError()
			}
			continue
		}

		if rule.isSlice {
			var s reflect.Value

			switch rule.t {
			case _BytesSlice:
				return nil, ErrGraphqlType
			case _IntSlice:
				switch rv := value.(type) {
				case []int64:
					if rule.vrange {
						for _, n := range rv {
							if !rule.checkVRangeByReflect(n) {
								return nil, rule.toFormError()
							}
						}
					}
					s = reflect.ValueOf(rv)
				default:
					return nil, ErrGraphqlType
				}
			case _UintSlice:
				switch rv := value.(type) {
				case []uint64:
					if rule.vrange {
						for _, n := range rv {
							if !rule.checkVRangeByReflect(n) {
								return nil, rule.toFormError()
							}
						}
					}
					s = reflect.ValueOf(rv)
				default:
					return nil, ErrGraphqlType
				}
			case _BoolSlice:
				switch rv := value.(type) {
				case []bool:
					s = reflect.ValueOf(rv)
				default:
					return nil, ErrGraphqlType
				}
			case _StringSlice:
				switch rv := value.(type) {
				case []string:
					if rule.lrange || rule.fn != nil || rule.reg != nil {
						for i, n := range rv {
							v, ok := rule.toBytes(gotils.S2B(n))
							if !ok {
								return nil, rule.toFormError()
							}
							rv[i] = string(v)
						}
					}
					s = reflect.ValueOf(rv)
				default:
					return nil, ErrGraphqlType
				}
			case _FloatSlice:
				switch rv := value.(type) {
				case []float64:
					if rule.vrange {
						for _, n := range rv {
							if !rule.checkVRangeByReflect(n) {
								return nil, rule.toFormError()
							}
						}
					}
					s = reflect.ValueOf(rv)
				default:
					return nil, ErrGraphqlType
				}
			}

			if s.IsValid() {
				if !rule.checkSizeRange(&s) {
					return nil, rule.toFormError()
				}
				field.Set(s)
			}
			continue
		}

		switch rule.t {
		case _Bytes:
			return nil, ErrGraphqlType
		case _String:
			var data []byte
			switch rv := value.(type) {
			case string:
				data = gotils.S2B(rv)
			default:
				return nil, ErrGraphqlType
			}
			v, ok := rule.toBytes(data)
			if !ok {
				return nil, rule.toFormError()
			}
			field.SetString(gotils.B2S(v))
		case _Uint64:
			switch rv := value.(type) {
			case uint64:
				if !rule.checkVRangeByReflect(value) {
					return nil, rule.toFormError()
				}
				field.SetUint(rv)
			case uint:
				v := uint64(rv)
				if !rule.checkVRangeByReflect(v) {
					return nil, rule.toFormError()
				}
				field.SetUint(v)
			default:
				return nil, ErrGraphqlType
			}
		case _Int64:
			switch rv := value.(type) {
			case int64:
				if !rule.checkVRangeByReflect(value) {
					return nil, rule.toFormError()
				}
				field.SetInt(rv)
			case int:
				v := int64(rv)
				if !rule.checkVRangeByReflect(v) {
					return nil, rule.toFormError()
				}
				field.SetInt(v)
			default:
				return nil, ErrGraphqlType
			}
		case _Float64:
			switch rv := value.(type) {
			case float64:
				if !rule.checkVRangeByReflect(value) {
					return nil, rule.toFormError()
				}
				field.SetFloat(rv)
			default:
				return nil, ErrGraphqlType
			}
		case _Bool:
			switch rv := value.(type) {
			case bool:
				field.SetBool(rv)
			default:
				return nil, ErrGraphqlType
			}
		default:
			return nil, ErrGraphqlType
		}
	}
	return &ele, nil
}
