package builder

import "github.com/zzztttkkk/sqrl"

type _logicT int

const (
	_AND = _logicT(iota)
	_OR
)

type _Conditions struct {
	v _logicT

	lt  sqrl.Lt
	gt  sqrl.Gt
	lte sqrl.LtOrEq
	gte sqrl.GtOrEq
	eq  sqrl.Eq
	neq sqrl.NotEq

	like   sqrl.Like
	unlike sqrl.NotLike

	ilike   sqrl.ILike
	unilike sqrl.NotILike
}

func (l *_Conditions) Lt(ok bool, key string, value interface{}) *_Conditions {
	if !ok {
		return l
	}
	if l.lt == nil {
		l.lt = sqrl.Lt{key: value}
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
		l.gt = sqrl.Gt{key: value}
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
		l.lte = sqrl.LtOrEq{key: value}
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
		l.gte = sqrl.GtOrEq{key: value}
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
		l.eq = sqrl.Eq{key: value}
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
		l.neq = sqrl.NotEq{key: value}
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
		l.like = sqrl.Like{key: value}
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
		l.unlike = sqrl.NotLike{key: value}
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
		l.ilike = sqrl.ILike{key: value}
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
		l.unilike = sqrl.NotILike{key: value}
	} else {
		l.unilike[key] = value
	}
	return l
}

func (l *_Conditions) ToSql() (string, []interface{}, error) {
	var sqlizers []sqrl.Sqlizer

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
		return sqrl.Or(sqlizers).ToSql()
	}
	return sqrl.And(sqlizers).ToSql()
}

func AndConditions() *_Conditions {
	return &_Conditions{v: _AND}
}

func OrConditions() *_Conditions {
	return &_Conditions{v: _OR}
}
