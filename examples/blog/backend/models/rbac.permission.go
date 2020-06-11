package models

import (
	"context"
	"github.com/zzztttkkk/snow/rbac"
	"github.com/zzztttkkk/snow/sqls"
)

type Permission struct {
	Id      uint32 `ddl:"notnull;incr;primary"`
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

func (op *_PermissionOperator) getAll(ctx context.Context) (lst []rbac.Permission) {
	op.SqlxStructScanRows(
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
	op.SqlxFetch(ctx, permission, `select * from permission where id=? and deleted=0`, id)
	return
}

func (op *_PermissionOperator) getByName(ctx context.Context, name string) (permission *Permission) {
	op.SqlxFetch(ctx, permission, `select * from permission where name=? and deleted=0`, name)
	return
}
