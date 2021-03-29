package jsonx

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

type JsonObj struct {
	v interface{}
	t reflect.Type

	peekAble bool
}

func (obj *JsonObj) String() string {
	return fmt.Sprintf("JsonObj(%s)", obj.v)
}

func NewObject(raw []byte) (*JsonObj, error) {
	var dist interface{}
	if err := Unmarshal(raw, &dist); err != nil {
		return nil, err
	}

	t := reflect.TypeOf(dist)
	obj := &JsonObj{v: dist, t: t}
	obj.peekAble = t.Kind() == reflect.Map || t.Kind() == reflect.Slice
	return obj, nil
}

var ErrUnexpectedJSON = errors.New("sha.jsonx: unexpected json structure")

const (
	SliceRand = "rand"
)

func peek(v interface{}, key string) (interface{}, error) {
	switch tv := v.(type) {
	case map[string]interface{}:
		rv, ok := tv[key]
		if ok {
			return rv, nil
		}
		return nil, ErrUnexpectedJSON
	case []interface{}:
		var ind int
		var l = len(tv)
		if key == SliceRand {
			ind = rand.Int() % l
		} else {
			iV, err := strconv.ParseInt(key, 10, 64)
			if err != nil || int(iV) > l {
				return nil, ErrUnexpectedJSON
			}
			if iV < 0 {
				iV = int64(l) + iV
			}
			if iV < 0 {
				return nil, ErrUnexpectedJSON
			}
			ind = int(iV)
		}
		return tv[ind], nil
	default:
		return nil, ErrUnexpectedJSON
	}
}

func (obj *JsonObj) Peek(keys ...string) (interface{}, error) {
	if len(keys) == 0 {
		return obj.v, nil
	}
	if !obj.peekAble {
		return nil, ErrUnexpectedJSON
	}

	v := obj.v
	var err error
	for _, key := range keys {
		v, err = peek(v, key)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (obj *JsonObj) PeekDefault(def interface{}, keys ...string) interface{} {
	v, e := obj.Peek(keys...)
	if e != nil {
		return def
	}
	return v
}

func (obj *JsonObj) PeekInt(keys ...string) (int64, error) {
	v, e := obj.Peek(keys...)
	if e != nil {
		return 0, e
	}

	switch tv := v.(type) {
	case float64:
		return int64(tv), nil
	default:
		return 0, ErrUnexpectedJSON
	}
}

func (obj *JsonObj) PeekIntDefault(def int64, keys ...string) int64 {
	v, e := obj.PeekInt(keys...)
	if e != nil {
		return def
	}
	return v
}

func (obj *JsonObj) PeekFloat(keys ...string) (float64, error) {
	v, e := obj.Peek(keys...)
	if e != nil {
		return 0, e
	}
	switch tv := v.(type) {
	case float64:
		return tv, nil
	default:
		return 0, ErrUnexpectedJSON
	}
}

func (obj *JsonObj) PeekFloatDefault(def float64, keys ...string) float64 {
	v, e := obj.PeekFloat(keys...)
	if e != nil {
		return def
	}
	return v
}

func (obj *JsonObj) PeekString(keys ...string) (string, error) {
	v, e := obj.Peek(keys...)
	if e != nil {
		return "", e
	}
	switch tv := v.(type) {
	case string:
		return tv, nil
	default:
		return "", ErrUnexpectedJSON
	}
}

func (obj *JsonObj) PeekStringDefault(def string, keys ...string) string {
	v, e := obj.PeekString(keys...)
	if e != nil {
		return def
	}
	return v
}

func (obj *JsonObj) PeekBool(keys ...string) (bool, error) {
	v, e := obj.Peek(keys...)
	if e != nil {
		return false, e
	}
	switch tv := v.(type) {
	case bool:
		return tv, nil
	default:
		return false, ErrUnexpectedJSON
	}
}

func (obj *JsonObj) PeekBoolDefault(def bool, keys ...string) bool {
	v, e := obj.PeekBool(keys...)
	if e != nil {
		return def
	}
	return v
}

func (obj *JsonObj) PeekMap(keys ...string) (map[string]interface{}, error) {
	v, e := obj.Peek(keys...)
	if e != nil {
		return nil, e
	}
	switch tv := v.(type) {
	case map[string]interface{}:
		return tv, nil
	default:
		return nil, ErrUnexpectedJSON
	}
}

func (obj *JsonObj) PeekSlice(keys ...string) ([]interface{}, error) {
	v, e := obj.Peek(keys...)
	if e != nil {
		return nil, e
	}
	switch tv := v.(type) {
	case []interface{}:
		return tv, nil
	default:
		return nil, ErrUnexpectedJSON
	}
}

func (obj *JsonObj) PeekTimeFromString(layout string, keys ...string) (time.Time, error) {
	v, err := obj.PeekString(keys...)
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(layout, v)
	if err != nil {
		return time.Time{}, ErrUnexpectedJSON
	}
	return t, nil
}

func (obj *JsonObj) PeekTimeFromUnix(keys ...string) (time.Time, error) {
	v, err := obj.PeekInt(keys...)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(v, 0), err
}

func (obj *JsonObj) IsNil(keys ...string) (bool, error) {
	v, err := obj.Peek(keys...)
	if err != nil {
		return false, err
	}
	return v == nil, nil
}
