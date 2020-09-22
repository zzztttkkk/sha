package jsonx

import (
	"database/sql/driver"
	"strconv"

	"github.com/savsgio/gotils"
)

type Array []interface{}

var _EmptyJsonArrayBytes = []byte("[]")

func (a Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return _EmptyJsonArrayBytes, nil
	}
	return Marshal(a)
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
	return Unmarshal(bytes, a)
}

func (a Array) get(key string) (interface{}, error) {
	ind, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return nil, ErrJsonValue
	}
	i := int(ind)
	if i < 0 || i >= len(a) {
		return nil, ErrJsonValue
	}
	return a[i], nil
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
	if err := Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (a Array) ToCollection() *Collection { return &Collection{raw: a} }
