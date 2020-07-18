package sqls

import (
	"context"
	"reflect"
	"sync"
)

type Operator struct {
	p   reflect.Type
	ddl *_DdlParser
}

func (op *Operator) Init(p reflect.Type) {
	op.p = p
	op.ddl = newDdlParser(p)

	config.SqlLeader().MustExec(TableDefinition(op.p))
}

func (op *Operator) TableName() string {
	return op.ddl.tableName
}

func (op *Operator) NewEnumCache(seconds int64, constructor func() Enumer) *EnumCache {
	cache := &EnumCache{
		im:          map[int64]interface{}{},
		nm:          map[string]interface{}{},
		last:        0,
		expire:      seconds,
		op:          op,
		constructor: constructor,
		rwm:         sync.RWMutex{},
	}
	cache.load(context.Background())
	return cache
}
