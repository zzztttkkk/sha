package models

import (
	"context"
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
)

type _EnumT struct {
	sqls.Model
	Name string `json:"name" ddl:"notnull;unique;L<30>"`
}

func (enum *_EnumT) GetId() int64 {
	return enum.Id
}

func (enum *_EnumT) GetName() string {
	return enum.Name
}

type _EnumOperatorT struct {
	sqls.Operator
	cache *sqls.EnumCache
}

func (op *_EnumOperatorT) Init(p reflect.Type, constructor func() sqls.Enum) {
	op.Operator.Init(p)
	op.SqlsTableCreate()
	op.cache = op.NewEnumCache(
		ini.GetIntOr("cache.sql.enum.expire", 1800),
		constructor,
	)
}

func (op *_EnumOperatorT) ExistsById(ctx context.Context, categoryId int64) bool {
	_, ok := op.cache.GetById(ctx, categoryId)
	return ok
}

func (op *_EnumOperatorT) ExistsByName(ctx context.Context, name string) bool {
	_, ok := op.cache.GetByName(ctx, name)
	return ok
}

func (op *_EnumOperatorT) List(ctx context.Context) []sqls.Enum {
	return op.cache.All(ctx)
}
