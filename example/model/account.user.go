package model

import (
	"context"
	"fmt"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna.example/config"
	"github.com/zzztttkkk/suna.example/internal"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/cache"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/sqls/builder"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"time"
)

type User struct {
	sqls.Model
	Name     string `json:"name"`
	Alias    string `json:"alias"`
	Password []byte `json:"-"`
	Secret   []byte `json:"-"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

func (User) TableName() string {
	return "account_user"
}

func (user User) TableDefinition() []string {
	return user.Model.TableDefinition(
		"name char(30) unique not null",
		"alias char(16) default ''",
		"password char(64) not null",
		"secret char(64) not null",
		"bio varchar(512) default ''",
		"avatar varchar(120) default ''",
	)
}

func (user *User) GetId() int64 {
	return user.Id
}

type _UserOperator struct {
	sqls.Operator
	lru *cache.Lru
}

var authTokenInHeader string
var authTokenInCookie string
var cfg *config.Example

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Example) {
			cfg = conf
			authTokenInHeader = conf.Auth.HeaderName
			authTokenInCookie = conf.Auth.CookieName
		},
	)
}

func readUid(ctx *fasthttp.RequestCtx) int64 {
	var token string
	if len(authTokenInHeader) > 0 {
		bytesV := ctx.Request.Header.Peek(authTokenInHeader)
		if len(bytesV) > 0 {
			token = gotils.B2S(bytesV)
		}
	}

	if len(token) < 1 && len(authTokenInCookie) > 1 {
		bytesV := ctx.Request.Header.Cookie(authTokenInCookie)
		if len(bytesV) > 0 {
			token = gotils.B2S(bytesV)
		}
	}

	if len(token) < 1 {
		return -1
	}

	id, ok := secret.LoadId(token)
	if !ok {
		return -1
	}
	return id
}

func (op *_UserOperator) Auth(ctx *fasthttp.RequestCtx) (auth.User, bool) {
	uid := readUid(ctx)
	if uid < 1 {
		return nil, false
	}
	return op.GetById(ctx, uid)
}

func (op *_UserOperator) Dump(uid int64, seconds int64) string { return secret.DumpId(uid, seconds) }

func (op *_UserOperator) GetById(ctx context.Context, uid int64) (*User, bool) {
	ui, ok := op.lru.Get(uid)
	if ok {
		return ui.(*User), false
	}

	user := &User{}
	op.XSelect(
		ctx,
		user,
		builder.NewSelect("*").
			From(op.TableName()).
			Where("id=? and status>=0 and deleted=0", uid).
			Limit(1),
	)
	if user.Id < 1 {
		return nil, false
	}
	op.lru.Add(uid, user)
	return user, true
}

func (op *_UserOperator) Create(ctx context.Context, name, password []byte) (int64, []byte) {
	if op.XExists(ctx, builder.AndConditions().Eq(true, "name", name)) {
		panic(output.NewError(fasthttp.StatusBadRequest, -1, fmt.Sprintf("`%s` is already taken.", name)))
	}
	skey := secret.Sha256.Calc(secret.RandBytes(512, nil))

	kvs := utils.AcquireKvs()
	defer kvs.Free()
	kvs.Append("created", time.Now().Unix())
	kvs.Append("name", name)
	kvs.Append("password", secret.Calc(password))
	kvs.Append("secret", skey)
	return op.XCreate(ctx, kvs), skey
}

func (op *_UserOperator) AuthByName(ctx context.Context, name, password []byte) (int64, bool) {
	var pwdHash []byte
	var uid int64

	op.XSelectScan(
		ctx,
		builder.NewSelect("id,password").From(op.TableName()).Where("name=?", name).Limit(1),
		sqls.NewScanner(
			[]interface{}{&uid, &pwdHash},
			nil,
			nil,
		),
	)

	if len(pwdHash) < 1 {
		return -1, false
	}
	return uid, secret.Equal(password, pwdHash)
}

func (op *_UserOperator) AuthById(ctx context.Context, id int64, password []byte) bool {
	var pwdHash []byte
	op.XSelect(
		ctx,
		&pwdHash,
		builder.NewSelect("password").From(op.TableName()).Where("id=?", id).Limit(1),
	)
	if len(pwdHash) < 1 {
		return false
	}
	return secret.Equal(password, pwdHash)
}

func (op *_UserOperator) Update(ctx context.Context, uid int64, kvs *utils.Kvs, secret []byte) bool {
	op.lru.Remove(uid)
	kvs.Remove("deleted")
	return op.XUpdate(
		ctx,
		kvs,
		builder.AndConditions().
			Eq(true, "id", uid).
			Eq(len(secret) > 0, "secret", secret),
		1,
	) > 0
}

func (op *_UserOperator) Delete(ctx context.Context, user *User, skey string) bool {
	op.lru.Remove(user.Id)
	kvs := utils.AcquireKvs()
	defer kvs.Free()
	kvs.Append("deleted", time.Now().Unix())
	return op.XUpdate(
		ctx,
		kvs,
		builder.AndConditions().
			Eq(true, "id", user.Id).
			Eq(true, "secret", skey),
		1,
	) > 0
}

var UserOperator = &_UserOperator{}

func init() {
	internal.LazyExecutor.Register(
		func(kwargs utils.Kwargs) {
			UserOperator.Init(reflect.ValueOf(User{}))
			UserOperator.lru = cache.NewLru(cfg.Cache.Lru.UserSize)
		},
	)
}
