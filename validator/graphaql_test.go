package validator

import (
	"fmt"
	"testing"
)

type A struct {
	Id int64
}

type B struct {
	Name string
}

type C struct {
	A
	B
	Text string   `validator:"txt:"`
	Aids []string `validator:"aids:"`
}

func TestRules_BindValue(t *testing.T) {
	v := GetRules(C{}).BindValue(
		map[string]interface{}{
			"id": 34, "name": "aaa",
			"txt": "----", "aids": []string{"a", "b"},
		},
	)
	fmt.Println(v)
}
