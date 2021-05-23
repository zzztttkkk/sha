package jsonx

import (
	json "github.com/goccy/go-json"
	"io"
)

func Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }

func Unmarshal(v []byte, d interface{}) error { return json.Unmarshal(v, d) }

func NewEncoder(w io.Writer) *json.Encoder { return json.NewEncoder(w) }
