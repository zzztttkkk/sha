package sqlx

import (
	"context"
	"github.com/zzztttkkk/suna/sqlx/sqlr"
)

func isPostgres(ctx context.Context, op *Operator) bool {
	db := GetLeaderDB(ctx)
	if db == nil {
		db = op._GetDbLeader()
	}
	return db.DriverName() == "postgres"
}

func (op *Operator) SelectBuilder(ctx context.Context, cols ...string) *sqlr.SelectBuilder {
	if isPostgres(ctx, op) {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Select(cols...)
	}
	return sqlr.Select(cols...)
}

func (op *Operator) UpdateBuilder(ctx context.Context, table string) *sqlr.UpdateBuilder {
	if isPostgres(ctx, op) {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Update(table)
	}
	return sqlr.Update(table)
}

func (op *Operator) DeleteBuilder(ctx context.Context, what ...string) *sqlr.DeleteBuilder {
	if isPostgres(ctx, op) {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Delete(what...)
	}
	return sqlr.Delete(what...)
}

func (op *Operator) InsertBuilder(ctx context.Context, table string) *sqlr.InsertBuilder {
	if isPostgres(ctx, op) {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Insert(table)
	}
	return sqlr.Insert(table)
}

type _logicT int

const (
	_AND = _logicT(iota)
	_OR
)

type _Conditions struct {
	v _logicT

	lt  sqlr.Lt
	gt  sqlr.Gt
	lte sqlr.LtOrEq
	gte sqlr.GtOrEq
	eq  sqlr.Eq
	neq sqlr.NotEq

	like    sqlr.Like
	notLike sqlr.NotLike

	iLike    sqlr.ILike
	notILike sqlr.NotILike
}

func (l *_Conditions) Lt(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.lt == nil {
		l.lt = sqlr.Lt{key: value}
	} else {
		l.lt[key] = value
	}

	return l
}

func (l *_Conditions) Gt(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.gt == nil {
		l.gt = sqlr.Gt{key: value}
	} else {
		l.gt[key] = value
	}

	return l
}

func (l *_Conditions) Lte(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.lte == nil {
		l.lte = sqlr.LtOrEq{key: value}
	} else {
		l.lte[key] = value
	}

	return l
}

func (l *_Conditions) Gte(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.gte == nil {
		l.gte = sqlr.GtOrEq{key: value}
	} else {
		l.gte[key] = value
	}

	return l
}

func (l *_Conditions) Eq(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.eq == nil {
		l.eq = sqlr.Eq{key: value}
	} else {
		l.eq[key] = value
	}

	return l
}

func (l *_Conditions) NotEq(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.neq == nil {
		l.neq = sqlr.NotEq{key: value}
	} else {
		l.neq[key] = value
	}

	return l
}

func (l *_Conditions) Like(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.like == nil {
		l.like = sqlr.Like{key: value}
	} else {
		l.like[key] = value
	}

	return l
}

func (l *_Conditions) NotLike(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.notLike == nil {
		l.notLike = sqlr.NotLike{key: value}
	} else {
		l.notLike[key] = value
	}
	return l
}

func (l *_Conditions) ILike(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.iLike == nil {
		l.iLike = sqlr.ILike{key: value}
	} else {
		l.iLike[key] = value
	}
	return l
}

func (l *_Conditions) NotILike(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.notILike == nil {
		l.notILike = sqlr.NotILike{key: value}
	} else {
		l.notILike[key] = value
	}
	return l
}

func (l *_Conditions) ToSql() (string, []interface{}, error) {
	var sqlizers []sqlr.Sqlizer

	if len(l.lt) > 0 {
		sqlizers = append(sqlizers, l.lt)
	}

	if len(l.gt) > 0 {
		sqlizers = append(sqlizers, l.gt)
	}

	if len(l.lte) > 0 {
		sqlizers = append(sqlizers, l.lte)
	}

	if len(l.gte) > 0 {
		sqlizers = append(sqlizers, l.gte)
	}

	if len(l.eq) > 0 {
		sqlizers = append(sqlizers, l.eq)
	}

	if len(l.neq) > 0 {
		sqlizers = append(sqlizers, l.neq)
	}

	if len(l.like) > 0 {
		sqlizers = append(sqlizers, l.like)
	}

	if len(l.notLike) > 0 {
		sqlizers = append(sqlizers, l.notLike)
	}

	if len(l.iLike) > 0 {
		sqlizers = append(sqlizers, l.iLike)
	}

	if len(l.notILike) > 0 {
		sqlizers = append(sqlizers, l.notILike)
	}

	if l.v == _OR {
		return sqlr.Or(sqlizers).ToSql()
	}
	return sqlr.And(sqlizers).ToSql()
}

func AND() *_Conditions {
	return &_Conditions{v: _AND}
}

func OR() *_Conditions {
	return &_Conditions{v: _OR}
}
