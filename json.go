package suna

import (
	"encoding/json"
	"io"
)

type JsonObject map[string]interface{}
type JsonArray []interface{}

var JsonMarshal func(v interface{}) ([]byte, error)
var JsonEncodeTo func(v interface{}, w io.Writer) error
var JsonDecodeFrom func(r io.Reader, dist interface{}) error
var JsonUnmarshal func(v []byte, dist interface{}) error
var JsonMustMarshal func(v interface{}) []byte

func init() {
	JsonMarshal = json.Marshal
	JsonEncodeTo = func(v interface{}, w io.Writer) error {
		encoder := json.NewEncoder(w)
		return encoder.Encode(v)
	}
	JsonDecodeFrom = func(r io.Reader, dist interface{}) error {
		decoder := json.NewDecoder(r)
		return decoder.Decode(dist)
	}
	JsonUnmarshal = json.Unmarshal
	JsonMustMarshal = func(v interface{}) []byte {
		d, e := JsonMarshal(v)
		if e != nil {
			panic(e)
		}
		return d
	}
}
