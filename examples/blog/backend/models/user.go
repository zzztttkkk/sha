package models

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
	"strconv"
	"time"
)

type User struct {
	sqls.Model
	Name     string `json:"name" ddl:"notnull;unique;L<30>"`
	Alias    string `json:"alias" ddl:"L<30>;D<''>"`
	Password []byte `json:"-" ddl:"notnull;L<64>"`
	Secret   []byte `json:"-" ddl:"notnull;L<64>"`
	Bio      string `json:"bio" ddl:"L<120>;D<''>"`
	Avatar   string `json:"avatar" ddl:"L<120>;D<''>"`

	roles string `json:"-" ddl:"-"`
}

func (user *User) GetId() int64 {
	return user.Id
}

func (user *User) GetRoleIds() string {
	return user.roles
}

type _UserOperatorT struct {
	sqls.Operator
}

var UserOperator = &_UserOperatorT{}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			UserOperator.Init(reflect.TypeOf(User{}))
		},
	)
}

func (op *_UserOperatorT) Create(ctx context.Context, name, password []byte) (int64, []byte) {
	if op.SqlxExists(ctx, `select count(id) from user where name=?`, name) {
		panic(output.NewHttpError(fasthttp.StatusBadRequest, -1, fmt.Sprintf("`%s` is already taken.", name)))
	}
	skey := secret.Sha256.Calc(secret.RandBytes(512, nil))

	return op.SqlxCreate(
		ctx,
		sqls.Dict{
			"created":  time.Now().Unix(),
			"name":     name,
			"password": secret.Hash.Calc(password),
			"secret":   skey,
		},
	), skey
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

	if len(pwdHash) < 1 {
		return -1, false
	}
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
	if len(pwdHash) < 1 {
		return false
	}
	return secret.Hash.Equal(password, pwdHash)
}

func (op *_UserOperatorT) Update(ctx context.Context, uid int64, dict sqls.Dict) bool {
	return op.SqlxUpdate(
		ctx,
		`update user set %s where id=?`,
		dict, uid,
	) > 0
}

func (op *_UserOperatorT) Delete(ctx context.Context, uid int64, skey string) bool {
	return op.SqlxUpdate(
		ctx,
		`update user set %s where id=? and secret=?`,
		sqls.Dict{"deleted": time.Now().Unix()}, uid, skey,
	) > 0
}

func (op *_UserOperatorT) GetById(ctx context.Context, uid int64) mware.User {
	user := &User{}
	op.SqlxFetchOne(ctx, user, `select * from user where id=? and deleted=0 and status>=0`, uid)
	if user.Id < 1 {
		return nil
	}

	var roles []uint32
	op.SqlxFetch(ctx, &roles, `select role from user_roles where user=? and deleted=0`, uid)

	last := len(roles) - 1
	for ind, rid := range roles {
		user.roles += strconv.FormatUint(uint64(rid), 10)
		if ind < last {
			user.roles += ","
		}
	}
	return user
}
