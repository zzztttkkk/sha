package jsonx

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJsonKey(t *testing.T) {
	v := _Key{}
	v.init(`...\\.`)
	for {
		k, ok := v.next()
		if !ok {
			break
		}
		fmt.Println("k", *k, len(*k))
	}
}

func TestJsonGet(t *testing.T) {
	data := `{"a": {"b.": [1, 2, 3, {"": "0.0"}]}}`
	var obj Object
	err := json.Unmarshal([]byte(data), &obj)
	if err != nil {
		panic(err)
	}
	v, err := obj.Get(`a.b\..3.`)
	if err != nil {
		panic(err)
	}
	fmt.Println(v.(string))
}

func TestParseJsonObject(t *testing.T) {
	fmt.Println(ParseObject(`{"a":34}`))
}
