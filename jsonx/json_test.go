package jsonx

import (
	"fmt"
	"testing"
)

func TestJsonGet(_ *testing.T) {
	data := `{"a": {"b.": [1, 2, 3, {"": "0.0", "c": null, "d": false}]}}`
	c := MustParse(data)
	fmt.Println(c.MustGet(`a.b\.`))
	fmt.Println(c.MustGet(`a.b\..3.`))
	fmt.Println(c.MustIsNull(`a.b\..3.c`))
	fmt.Println(string(MustStringify(c)))
}
