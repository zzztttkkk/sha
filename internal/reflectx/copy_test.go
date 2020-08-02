package reflectx

import (
	"reflect"
	"testing"
)

type A struct {
	B struct {
		Name string
	}
}

func TestCopy(t *testing.T) {
	a := &A{}
	b := &A{}
	b.B.Name = "SS"

	Copy(reflect.ValueOf(a), reflect.ValueOf(b))
}
