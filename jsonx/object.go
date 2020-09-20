package jsonx

import (
	"database/sql/driver"
	"github.com/savsgio/gotils"
)

type Object map[string]interface{}

var emptyJsonObjBytes = []byte("{}")

func (f Object) Value() (driver.Value, error) {
	if len(f) == 0 {
		return emptyJsonObjBytes, nil
	}
	return Marshal(f)
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

func (f Object) Len() int {
	return len(f)
}

func (f Object) Get(key string) (interface{}, error) {
	return get(f, key)
}

func (f Object) MustGet(key string) interface{} {
	v, e := f.Get(key)
	if e != nil {
		panic(e)
	}
	return v
}

func (f Object) GetInt(key string) (int64, error) {
	return getInt64(f, key)
}

func (f Object) MustGetInt(key string) int64 {
	v, e := getInt64(f, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (f Object) GetFloat(key string) (float64, error) {
	return getFloat(f, key)
}

func (f Object) MustGetFloat(key string) float64 {
	v, e := getFloat(f, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (f Object) GetBool(key string) (bool, error) {
	return getBool(f, key)
}

func (f Object) MustGetBool(key string) bool {
	v, e := getBool(f, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (f Object) GetString(key string) (string, error) {
	return getString(f, key)
}

func (f Object) MustGetString(key string) string {
	v, e := getString(f, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (f Object) IsNull(key string) (bool, error) {
	return isNull(f, key)
}

func (f Object) MustIsNull(key string) bool {
	v, e := isNull(f, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (f Object) Set(key string, val interface{}) error {
	return set(f, key, val)
}

func (f Object) MustSet(key string, val interface{}) {
	if err := f.Set(key, val); err != nil {
		panic(err)
	}
}
