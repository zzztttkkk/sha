package models

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
	"time"
)

type _UserOperatorT struct {
	sqls.Operator
}

var UserOperator = &_UserOperatorT{}

func (op *_UserOperatorT) Create(ctx context.Context, name, password []byte) int64 {
	if op.SqlxExists(ctx, `select count(id) from user where name=?`, name) {
		panic(output.StdErrors[fasthttp.StatusBadRequest])
	}
	return op.SqlxCreate(
		ctx,
		`insert into user (created,name,password) values (?,?,?)`,
		time.Now().Unix(),
		name,
		secret.Hash.Calc(password),
	)
}

func (op *_UserOperatorT) AuthByName(ctx context.Context, name, password []byte) (int64, bool) {
	var pwdHash []byte
	var uid int64
	op.SqlxFetchOne(
		ctx,
		[]interface{}{&uid, &password},
		`select id,password from user where name=? and deleted=0 and status>-1`,
		name,
	)
	return uid, secret.Hash.Equal(password, pwdHash)
}

func (op *_UserOperatorT) AuthById(ctx context.Context, id int64, password []byte) bool {
	var pwdHash []byte
	op.SqlxFetchOne(
		ctx,
		&pwdHash,
		`select password from user where id=? and deleted=0 and status>-1`,
		id,
	)
	return secret.Hash.Equal(password, pwdHash)
}
