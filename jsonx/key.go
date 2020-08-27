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
		ind, err := s2i4(key)
		if err != nil {
			return nil, err
		}
		return rv.get(ind)
	case []interface{}:
		ind, err := s2i4(key)
		if err != nil {
			return nil, err
		}
		return Array(rv).get(ind)
	case Object:
		return rv.get(key)
	case map[string]interface{}:
		return Object(rv).get(key)
	default:
		return nil, ErrJsonValue
	}
}
