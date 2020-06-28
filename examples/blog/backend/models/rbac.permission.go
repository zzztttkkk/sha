package models

import (
	"context"
	"reflect"
	"time"

	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/rbac"
	"github.com/zzztttkkk/snow/sqls"
)

type Permission struct {
	Id      uint32 `ddl:"notnull;incr;primary" json:"id"`
	Name    string `ddl:"notnull;unique;T<char(255)>" json:"name"`
	Created int64  `ddl:"notnull" json:"created"`
	Deleted int64  `ddl:"D<0>" json:"deleted"`
	Descp   string `ddl:"L<0>" json:"descp"`
}

func (p *Permission) GetName() string {
	return p.Name
}

func (p *Permission) GetId() uint32 {
	return p.Id
}

type _PermissionOperator struct {
	sqls.Operator
}

var permissionOp = &_PermissionOperator{}

func init() {
	internal.LazyExecutor.RegisterWithPriority(
		func(kwargs snow.Kwargs) {
			permissionOp.Init(reflect.TypeOf(Permission{}))
		},
		permissionOpInitPriority,
	)
}

func (op *_PermissionOperator) getAll(ctx context.Context) (lst []rbac.Permission) {
	op.XStructScanMany(
		ctx,
		func() interface{} {
			permission := &Permission{}
			lst = append(lst, permission)
			return permission
		},
		"select * from permission where deleted=0",
	)
	return
}

func (op *_PermissionOperator) getById(ctx context.Context, id uint32) (permission *Permission) {
	op.XFetchMany(ctx, permission, `select * from permission where id=? and deleted=0`, id)
	return
}

func (op *_PermissionOperator) getByName(ctx context.Context, name string) (permission *Permission) {
	op.XFetchMany(ctx, permission, `select * from permission where name=? and deleted=0`, name)
	return
}

func (op *_PermissionOperator) createIfNotExists(ctx context.Context, name, descp string) {
	if op.XExists(ctx, "select count(id) from permission where name=? and deleted=0", name) {
		return
	}
	op.XCreate(ctx, sqls.Dict{"name": name, "created": time.Now().Unix(), "descp": descp})
}
