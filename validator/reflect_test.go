package validator

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValidate(t *testing.T) {
	type Form struct {
		Password  []byte `vld:":R<password>"`
		Name      []byte `vld:":F<username>"`
		KeepLogin bool   `vld:"kl:optional"`
	}
	fmt.Println(getRules(reflect.TypeOf(Form{})))
}
