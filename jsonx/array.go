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

func (a Array) set(key string, val interface{}) error {
	ind, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return ErrJsonValue
	}
	i := int(ind)
	if i < 0 || i >= len(a) {
		return ErrJsonValue
	}
	a[i] = val
	return nil
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

func (a Array) Len() int {
	return len(a)
}

func (a Array) Get(key string) (interface{}, error) {
	return get(a, key)
}

func (a Array) MustGet(key string) interface{} {
	v, e := a.Get(key)
	if e != nil {
		panic(e)
	}
	return v
}

func (a Array) GetInt(key string) (int64, error) {
	return getInt64(a, key)
}

func (a Array) MustGetInt(key string) int64 {
	v, e := getInt64(a, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (a Array) GetFloat(key string) (float64, error) {
	return getFloat(a, key)
}

func (a Array) MustGetFloat(key string) float64 {
	v, e := getFloat(a, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (a Array) GetBool(key string) (bool, error) {
	return getBool(a, key)
}

func (a Array) MustGetBool(key string) bool {
	v, e := getBool(a, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (a Array) GetString(key string) (string, error) {
	return getString(a, key)
}

func (a Array) MustGetString(key string) string {
	v, e := getString(a, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (a Array) IsNull(key string) (bool, error) {
	return isNull(a, key)
}

func (a Array) MustIsNull(key string) bool {
	v, e := isNull(a, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (a Array) Set(key string, val interface{}) error {
	return set(a, key, val)
}

func (a Array) MustSet(key string, val interface{}) {
	if err := a.Set(key, val); err != nil {
		panic(err)
	}
}
