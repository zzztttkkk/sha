package models

import (
	"context"
	"database/sql"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/rbac"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
	"strconv"
)

type Role struct {
	Id          uint32 `ddl:"notnull;incr;primary"`
	Name        string `ddl:"notnull;unique;T<char(255)>" json:"name"`
	Created     int64  `ddl:"notnull" json:"created"`
	Deleted     int64  `ddl:"D<0>" json:"deleted"`
	Descp       string `ddl:"L<0>" json:"descp"`
	Parent      uint32 `ddl:"D<0>" json:"parent"`
	Permissions string `json:"permissions" ddl:"-"`
}

func (role *Role) GetParentId() uint32 {
	return role.Parent
}

func (role *Role) GetName() string {
	return role.Name
}

func (role *Role) GetId() uint32 {
	return role.Id
}

func (role *Role) GetPermissionIds() string {
	return role.Permissions
}

type _RoleOperator struct {
	sqls.Operator
}

var roleOp = &_RoleOperator{}

func init() {
	internal.LazyExecutor.RegisterWithPriority(
		func(kwargs snow.Kwargs) {
			roleOp.Init(reflect.TypeOf(Role{}))
		},
		permissionPriority.Copy(),
	)
}

func (op *_RoleOperator) getAll(ctx context.Context) (lst []rbac.Role) {
	var roles []*Role
	op.SqlxStructScanRows(
		ctx,
		func() interface{} {
			role := &Role{}
			lst = append(lst, role)
			roles = append(roles, role)
			return role
		},
		"select * from role where deleted=0",
	)

	stmt := op.SqlxStmt(ctx, `select distinct(permission) from role_permissions where deleted=0 and role=?`)
	defer stmt.Close()
	var allRows []*sql.Rows
	defer func() {
		for _, rows := range allRows {
			_ = rows.Close()
		}
	}()

	var pid uint64
	for _, role := range roles {
		rows, err := stmt.Query(role.Id)
		if err != nil {
			panic(err)
		}
		allRows = append(allRows, rows)

		for rows.Next() {
			_ = rows.Scan(&pid)
			role.Permissions += strconv.FormatUint(pid, 10) + ","
		}

		if role.Permissions[len(role.Permissions)-1] == ',' {
			role.Permissions = role.Permissions[:len(role.Permissions)-1]
		}
	}
	return
}

func (op *_RoleOperator) getById(ctx context.Context, id uint32) (role *Role) {
	op.SqlxFetch(ctx, role, `select * from role where id=? and deleted=0`, id)
	return
}

func (op *_RoleOperator) getByName(ctx context.Context, name string) (role *Role) {
	op.SqlxFetch(ctx, role, `select * from role where name=? and deleted=0`, name)
	return
}
