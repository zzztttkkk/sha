package builder

import "github.com/zzztttkkk/sqlr"

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

	like   sqlr.Like
	unlike sqlr.NotLike

	ilike   sqlr.ILike
	unilike sqlr.NotILike
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
	if l.unlike == nil {
		l.unlike = sqlr.NotLike{key: value}
	} else {
		l.unlike[key] = value
	}
	return l
}

func (l *_Conditions) ILike(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.ilike == nil {
		l.ilike = sqlr.ILike{key: value}
	} else {
		l.ilike[key] = value
	}
	return l
}

func (l *_Conditions) NotILike(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.unilike == nil {
		l.unilike = sqlr.NotILike{key: value}
	} else {
		l.unilike[key] = value
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

	if len(l.unlike) > 0 {
		sqlizers = append(sqlizers, l.unlike)
	}

	if len(l.ilike) > 0 {
		sqlizers = append(sqlizers, l.ilike)
	}

	if len(l.unilike) > 0 {
		sqlizers = append(sqlizers, l.unilike)
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
