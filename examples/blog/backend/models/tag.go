package models

import (
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
)

type Tag struct {
	_EnumT
	Descp string `json:"descp" ddl:"L<0>;notnull"`
}

type _PostTags struct {
	TagId   int64 `ddl:"primary"`
	PostId  int64 `ddl:"primary"`
	Created int64 `ddl:"notnull"`
	Deleted int64 `ddl:"D<0>"`
}

type _TagOperatorT struct {
	_EnumOperatorT
}

var TagOperator = &_TagOperatorT{}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			TagOperator.Init(
				reflect.TypeOf(Tag{}),
				func() sqls.Enum { return &Tag{} },
			)
		},
	)
}
