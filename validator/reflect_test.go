package validator

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValidate(t *testing.T) {
	type Form struct {
		Password  []byte `validator:":R<password>"`
		Name      []byte `validator:":F<username>"`
		KeepLogin bool   `validator:"kl:optional"`
	}
	fmt.Println(getRules(reflect.TypeOf(Form{})))
}
