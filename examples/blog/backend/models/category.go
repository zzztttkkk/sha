package models

import (
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
)

type Category struct {
	sqls.Enum
	Descp string `json:"descp" ddl:"L<0>;notnull"`
}

type _CategoryOperatorT struct {
	sqls.EnumOperator
}

var CategoryOperator = &_CategoryOperatorT{}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			CategoryOperator.Init(
				reflect.TypeOf(Category{}),
				func() sqls.Enumer { return &Category{} },
			)
		},
	)
}
