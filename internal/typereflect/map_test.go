package typereflect

import (
	"fmt"
	"reflect"
	"testing"
)

type A struct {
	A int `json:"a"`
	B int `json:"b"`
}

type C struct {
	A A
	D int `json:"d"`
}

type E struct {
	F int `json:"f"`
	G int `json:"g"`
}

type K struct {
	E E
	C

}

func TestMap(t *testing.T) {
	fmt.Println(Keys(reflect.TypeOf(K{}), "json"))
}
