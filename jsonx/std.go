package jsonx

import (
	jsoniter "github.com/json-iterator/go"
	"io"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }

func Unmarshal(v []byte, d interface{}) error { return json.Unmarshal(v, d) }

func NewEncoder(w io.Writer) *jsoniter.Encoder { return jsoniter.NewEncoder(w) }

func GetOnce(data []byte, path ...interface{}) jsoniter.Any { return json.Get(data, path...) }
