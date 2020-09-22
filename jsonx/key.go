package jsonx

import (
	"reflect"
	"strings"
	"unsafe"
)

type _Key struct {
	raw          string
	buf          []byte
	cursor       int
	end          int
	str          reflect.SliceHeader
	lastEmptyFix bool
}

func (key *_Key) init(v string) {
	key.raw = v
	key.end = len(key.raw) - 1

	if key.raw[key.end] == '.' {
		if strings.HasSuffix(key.raw, `\.`) {
			if strings.HasSuffix(key.raw, `\\.`) {
				key.lastEmptyFix = true
			} else {
				key.lastEmptyFix = false
			}
		} else {
			key.lastEmptyFix = true
		}
	}
}

func (key *_Key) getStr() *string {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&key.buf))
	key.str.Data = sh.Data
	key.str.Len = sh.Len
	key.str.Cap = sh.Len
	return (*string)(unsafe.Pointer(&key.str))
}

func (key *_Key) next() (*string, bool) {
	key.buf = key.buf[:0]
	if key.cursor > key.end {
		if key.lastEmptyFix {
			key.lastEmptyFix = false
			return key.getStr(), true
		}
		return nil, false
	}

	var prev byte
	var b byte

	for ; key.cursor <= key.end; key.cursor++ {
		b = key.raw[key.cursor]

		if prev == '\\' {
			key.buf = append(key.buf, b)
			prev = 0
		} else {
			if b == '\\' {
				prev = b
			} else {
				if b == '.' {
					key.cursor++
					break
				}
				key.buf = append(key.buf, b)
			}
		}
	}

	return key.getStr(), true
}

func getFromInterface(key string, v interface{}) (interface{}, error) {
	switch rv := v.(type) {
	case Array:
		return rv.get(key)
	case []interface{}:
		return Array(rv).get(key)
	case Object:
		return rv.get(key)
	case map[string]interface{}:
		return Object(rv).get(key)
	default:
		return nil, ErrJsonValue
	}
}

type _JsonCollection interface {
	get(string) (interface{}, error)
}

func get(collection _JsonCollection, key string) (interface{}, error) {
	if len(key) < 1 {
		return collection, nil
	}

	k := _Key{}
	k.init(key)

	var rv interface{} = collection
	var err error
	var _k *string
	var ok bool
	for {
		_k, ok = k.next()
		if !ok {
			break
		}
		rv, err = getFromInterface(*_k, rv)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

func getInt64(collection _JsonCollection, key string) (int64, error) {
	v, err := get(collection, key)
	if err != nil {
		return 0, err
	}
	switch rv := v.(type) {
	case float64:
		return int64(rv), nil
	}
	return 0, ErrJsonValue
}

func getString(collection _JsonCollection, key string) (string, error) {
	v, err := get(collection, key)
	if err != nil {
		return "", err
	}
	switch rv := v.(type) {
	case string:
		return rv, nil
	}
	return "", ErrJsonValue
}

func getBool(collection _JsonCollection, key string) (bool, error) {
	v, err := get(collection, key)
	if err != nil {
		return false, err
	}
	switch rv := v.(type) {
	case bool:
		return rv, nil
	}
	return false, ErrJsonValue
}

func getFloat(collection _JsonCollection, key string) (float64, error) {
	v, err := get(collection, key)
	if err != nil {
		return 0, err
	}
	switch rv := v.(type) {
	case float64:
		return rv, nil
	}
	return 0, ErrJsonValue
}

func isNull(collection _JsonCollection, key string) (bool, error) {
	v, err := get(collection, key)
	if err != nil {
		return false, err
	}
	return v == nil, nil
}
