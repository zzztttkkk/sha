package jsonx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/savsgio/gotils"
	"strconv"
)

type Array []interface{}

var _EmptyJsonArrayBytes = []byte("[]")

func (a Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return _EmptyJsonArrayBytes, nil
	}
	v, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (a *Array) Scan(src interface{}) error {
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = gotils.S2B(v)
	case []byte:
		bytes = v
	default:
		return ErrJsonValue
	}
	return json.Unmarshal(bytes, a)
}

func (a Array) get(key int) (interface{}, error) {
	if key < 0 || key >= len(a) {
		return nil, ErrJsonValue
	}
	return a[key], nil
}

var ErrJsonValue = errors.New("suna.jsonx: json type error")

func s2i4(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrJsonValue
	}
	return int(v), nil
}

func (a Array) Get(key string) (interface{}, error) {
	k := _Key{}
	k.init(key)

	var rv interface{} = a
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

func (a Array) GetMust(key string) interface{} {
	v, e := a.Get(key)
	if e != nil {
		panic(e)
	}
	return v
}

func ParseArray(v interface{}) (Array, error) {
	var data []byte
	switch rv := v.(type) {
	case string:
		data = gotils.S2B(rv)
	case *string:
		if rv == nil {
			data = nil
		} else {
			data = gotils.S2B(*rv)
		}
	case []byte:
		data = rv
	case *[]byte:
		if rv == nil {
			data = nil
		} else {
			data = *rv
		}
	default:
		return nil, ErrJsonValue
	}
	m := Array{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
