package utils

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/savsgio/gotils"
)

type JsonObject map[string]interface{}

var emptyJsonObjBytes = []byte("{}")

func (f JsonObject) Value() (driver.Value, error) {
	if len(f) == 0 {
		return emptyJsonObjBytes, nil
	}
	v, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (f *JsonObject) Scan(src interface{}) error {
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = gotils.S2B(v)
	case []byte:
		bytes = v
	default:
		return ErrJsonValue
	}
	return json.Unmarshal(bytes, f)
}

func (f JsonObject) get(key string) (interface{}, error) {
	v, ok := f[key]
	if !ok {
		return nil, ErrJsonValue
	}
	return v, nil
}

func (f JsonObject) Get(key string) (interface{}, error) {
	k := _JsonKey{}
	k.init(key)

	var rv interface{} = f
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

func (f JsonObject) MustGet(key string) interface{} {
	v, e := f.Get(key)
	if e != nil {
		panic(e)
	}
	return v
}
