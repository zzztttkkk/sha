package sqlu

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/zzztttkkk/suna/utils"
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

var IsNotSqlJsonFieldError = errors.New("suna.sqlu: this is not a json field")

func (f *JsonObject) Scan(src interface{}) error {
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = utils.S2b(v)
	case []byte:
		bytes = v
	default:
		return IsNotSqlJsonFieldError
	}
	return json.Unmarshal(bytes, f)
}

type JsonArray []interface{}

var emptyJsonArrayBytes = []byte("[]")

func (a JsonArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return emptyJsonArrayBytes, nil
	}
	v, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (a *JsonArray) Scan(src interface{}) error {
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = utils.S2b(v)
	case []byte:
		bytes = v
	default:
		return IsNotSqlJsonFieldError
	}
	return json.Unmarshal(bytes, a)
}
