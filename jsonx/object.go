package jsonx

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/savsgio/gotils"
)

type Object map[string]interface{}

var emptyJsonObjBytes = []byte("{}")

func (f Object) Value() (driver.Value, error) {
	if len(f) == 0 {
		return emptyJsonObjBytes, nil
	}
	v, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (f *Object) Scan(src interface{}) error {
	m, e := ParseObject(src)
	if e != nil {
		return e
	}
	*f = m
	return nil
}

func (f Object) get(key string) (interface{}, error) {
	v, ok := f[key]
	if !ok {
		return nil, ErrJsonValue
	}
	return v, nil
}

func (f Object) set(key string, val interface{}) error {
	f[key] = val
	return nil
}

func ParseObject(v interface{}) (Object, error) {
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
	m := Object{}
	if err := Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
