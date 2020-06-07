package internal

import (
	"fmt"
	"reflect"
	"testing"
)

type _a struct {
	Z int
}

type _v struct {
	X int
}

type _E struct {
	_a
	_v `ddl:"-"`
	A  int `ddl:"aid:notnull;primary;Default<12>;S<>"`
	B  int `ddl:":"`
	C  int `ddl:":ccc"`
	D  int `ddl:"ccc:"`
}

type _Parser struct{}

func (p *_Parser) Tag() string {
	return "ddl"
}

func (p *_Parser) OnField(f *reflect.StructField) bool {
	if f.Type.Kind() != reflect.Int {
		return false
	}
	return true
}

func (p *_Parser) OnName(name string) {
	fmt.Println(name)
}

func (p *_Parser) OnAttr(key, val string) {
	fmt.Println(key, val)
}

func (p *_Parser) OnDone() {
}

func TestTagMap(t *testing.T) {
	ReflectTags(
		reflect.TypeOf(_E{}),
		&_Parser{},
	)
}
