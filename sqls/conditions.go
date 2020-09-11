package sqls

import (
	ci "github.com/zzztttkkk/suna/sqls/internal"
)

type _logicT int

const (
	_AND = _logicT(iota)
	_OR
)

type _OpConditions struct {
	v _logicT

	lt  ci.Lt
	gt  ci.Gt
	lte ci.LtOrEq
	gte ci.GtOrEq
	eq  ci.Eq
	neq ci.NotEq

	like    ci.Like
	notLike ci.NotLike

	iLike    ci.ILike
	notILike ci.NotILike
}

func (oc *_OpConditions) Lt(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.lt == nil {
		oc.lt = ci.Lt{key: value}
	} else {
		oc.lt[key] = value
	}

	return oc
}

func (oc *_OpConditions) Gt(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.gt == nil {
		oc.gt = ci.Gt{key: value}
	} else {
		oc.gt[key] = value
	}

	return oc
}

func (oc *_OpConditions) Lte(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.lte == nil {
		oc.lte = ci.LtOrEq{key: value}
	} else {
		oc.lte[key] = value
	}

	return oc
}

func (oc *_OpConditions) Gte(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.gte == nil {
		oc.gte = ci.GtOrEq{key: value}
	} else {
		oc.gte[key] = value
	}

	return oc
}

func (oc *_OpConditions) Eq(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.eq == nil {
		oc.eq = ci.Eq{key: value}
	} else {
		oc.eq[key] = value
	}

	return oc
}

func (oc *_OpConditions) NotEq(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.neq == nil {
		oc.neq = ci.NotEq{key: value}
	} else {
		oc.neq[key] = value
	}

	return oc
}

func (oc *_OpConditions) Like(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.like == nil {
		oc.like = ci.Like{key: value}
	} else {
		oc.like[key] = value
	}

	return oc
}

func (oc *_OpConditions) NotLike(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.notLike == nil {
		oc.notLike = ci.NotLike{key: value}
	} else {
		oc.notLike[key] = value
	}
	return oc
}

func (oc *_OpConditions) ILike(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.iLike == nil {
		oc.iLike = ci.ILike{key: value}
	} else {
		oc.iLike[key] = value
	}
	return oc
}

func (oc *_OpConditions) NotILike(ok bool, key string, value interface{}) *_OpConditions {
	if !ok {
		return oc
	}
	if oc.notILike == nil {
		oc.notILike = ci.NotILike{key: value}
	} else {
		oc.notILike[key] = value
	}
	return oc
}

func (oc *_OpConditions) ToSql() (string, []interface{}, error) {
	var sqlizers []ci.Sqlizer

	if len(oc.lt) > 0 {
		sqlizers = append(sqlizers, oc.lt)
	}

	if len(oc.gt) > 0 {
		sqlizers = append(sqlizers, oc.gt)
	}

	if len(oc.lte) > 0 {
		sqlizers = append(sqlizers, oc.lte)
	}

	if len(oc.gte) > 0 {
		sqlizers = append(sqlizers, oc.gte)
	}

	if len(oc.eq) > 0 {
		sqlizers = append(sqlizers, oc.eq)
	}

	if len(oc.neq) > 0 {
		sqlizers = append(sqlizers, oc.neq)
	}

	if len(oc.like) > 0 {
		sqlizers = append(sqlizers, oc.like)
	}

	if len(oc.notLike) > 0 {
		sqlizers = append(sqlizers, oc.notLike)
	}

	if len(oc.iLike) > 0 {
		sqlizers = append(sqlizers, oc.iLike)
	}

	if len(oc.notILike) > 0 {
		sqlizers = append(sqlizers, oc.notILike)
	}

	if oc.v == _OR {
		return ci.Or(sqlizers).ToSql()
	}
	return ci.And(sqlizers).ToSql()
}

func AND() *_OpConditions {
	return &_OpConditions{v: _AND}
}

func OR() *_OpConditions {
	return &_OpConditions{v: _OR}
}

type _StrConditions struct {
	s    string
	args []interface{}
}

func (sc *_StrConditions) ToSql() (string, []interface{}, error) {
	return sc.s, sc.args, nil
}

func STR(s string, args ...interface{}) *_StrConditions {
	return &_StrConditions{s: s, args: args}
}

func RAW(raw string) ci.Raw {
	return ci.Raw(raw)
}
