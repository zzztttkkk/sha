package jsonx

import (
	"encoding/json"
	"io"
)

type EncodeOption struct {
	EscapeHTML bool
	Prefix     string
	Ident      string
}

var Unmarshal = json.Unmarshal
var Marshal = json.Marshal
var EncodeTo = func(v interface{}, w io.Writer, option *EncodeOption) error {
	encoder := json.NewEncoder(w)
	if option != nil {
		encoder.SetEscapeHTML(option.EscapeHTML)
		encoder.SetIndent(option.Prefix, option.Ident)
	}
	return encoder.Encode(v)
}

func MustMarshal(v interface{}) []byte {
	b, e := Marshal(v)
	if e != nil {
		panic(e)
	}
	return b
}
