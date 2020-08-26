package grafql

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
	"testing"
)

type A struct {
	Id int64
}

func (A) ResolveId(p graphql.ResolveParams) (interface{}, error) {
	return nil, nil
}

type B struct {
	A
	Name string
}

func TestNewObject(t *testing.T) {
	v := NewOutObject(reflect.ValueOf(B{}))
	fmt.Println(v)
}
