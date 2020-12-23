package sqlx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/zzztttkkk/sha/utils"
)

type JsonBytesString []byte

var emptyString = []byte("\"\"")

func (u JsonBytesString) MarshalJSON() ([]byte, error) {
	if u == nil {
		return emptyString, nil
	}

	var ret []byte
	ret = append(ret, '"')

	for _, v := range u {
		switch v {
		case '"':
			ret = append(ret, '\\', v)
		case '\\':
			ret = append(ret, '\\', '\\')
		default:
			ret = append(ret, v)
		}
	}
	ret = append(ret, '"')
	return ret, nil
}

type JsonArray []interface{}

var emptyJsonArrayBytes = []byte("[]")

var ErrJsonValue = errors.New("sha.sqlx: json value error")

func (a JsonArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return emptyJsonArrayBytes, nil
	}
	return json.Marshal(a)
}

func (a *JsonArray) Scan(src interface{}) error {
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = utils.B(v)
	case []byte:
		bytes = v
	case *string:
		if v != nil {
			bytes = utils.B(*v)
		}
	case *[]byte:
		if v != nil {
			bytes = *v
		}
	default:
		return ErrJsonValue
	}
	return json.Unmarshal(bytes, a)
}

type JsonObject map[string]interface{}

var emptyJsonObjBytes = []byte("{}")

func (f JsonObject) Value() (driver.Value, error) {
	if len(f) == 0 {
		return emptyJsonObjBytes, nil
	}
	return json.Marshal(f)
}

func (f *JsonObject) Scan(src interface{}) error {
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = utils.B(v)
	case []byte:
		bytes = v
	case *string:
		if v != nil {
			bytes = utils.B(*v)
		}
	case *[]byte:
		if v != nil {
			bytes = *v
		}
	default:
		return ErrJsonValue
	}
	return json.Unmarshal(bytes, f)
}
