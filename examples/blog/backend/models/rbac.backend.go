package models

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/rbac"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
	"sync"
	"time"
)

type _BackendT struct {
	permissionOp *_PermissionOperator
	roleOp       *_RoleOperator
	mutex        sync.Mutex
}

func (backend *_BackendT) Changed(ctx context.Context) bool {
	return false
}

func (backend *_BackendT) LoadDone(ctx context.Context) {

}

func (backend *_BackendT) GetAllPermissions(ctx context.Context) []rbac.Permission {
	return backend.permissionOp.getAll(ctx)
}

func (backend *_BackendT) GetAllRoles(ctx context.Context) []rbac.Role {
	return backend.roleOp.getAll(ctx)
}

func (backend *_BackendT) AddPermission(ctx context.Context, permission string) {
	backend.permissionOp.SqlxCreate(
		ctx,
		sqls.Dict{"name": permission, "created": time.Now().Unix()},
	)
}

func (backend *_BackendT) DelPermission(ctx context.Context, permission string) {
	backend.permissionOp.SqlxUpdate(
		ctx,
		`update permission set deleted=? where name=?`,
		time.Now().Unix(), permission,
	)
}

func (backend *_BackendT) AddRole(ctx context.Context, role string) {
	backend.roleOp.SqlxCreate(
		ctx,
		sqls.Dict{"name": role, "created": time.Now().Unix()},
	)
}

func (backend *_BackendT) DelRole(ctx context.Context, role string) {
	backend.permissionOp.SqlxUpdate(
		ctx,
		`update role set deleted=? where name=?`,
		time.Now().Unix(), role,
	)
}

func (backend *_BackendT) AddPermissionForRole(ctx context.Context, role, permission string) {
	r := backend.roleOp.getByName(ctx, role)
	if r == nil {
		panic(fmt.Errorf("nil role, `%s`", role))
	}
	p := backend.permissionOp.getByName(ctx, permission)
	if p == nil {
		panic(fmt.Errorf("nil permission, `%s`", permission))
	}

	backend.permissionOp.SqlxExecute(
		ctx,
		`insert into role_permissions (created,role,permission) values(?,?,?)`,
		time.Now().Unix(), r.Id, p.Id,
	)
}

func (backend *_BackendT) DelPermissionForRole(ctx context.Context, role, permission string) {
	r := backend.roleOp.getByName(ctx, role)
	if r == nil {
		panic(fmt.Errorf("nil role, `%s`", role))
	}
	p := backend.permissionOp.getByName(ctx, permission)
	if p == nil {
		panic(fmt.Errorf("nil permission, `%s`", permission))
	}

	backend.permissionOp.SqlxExecute(
		ctx,
		`update role_permissions set deleted=? where role=? and permission=? and deleted=0`,
		time.Now().Unix(), r.Id, p.Id,
	)
}

func (backend *_BackendT) AddRoleForUser(ctx context.Context, role string, uid int64) {
	r := backend.roleOp.getByName(ctx, role)
	if r == nil {
		panic(fmt.Errorf("nil role, `%s`", role))
	}

	backend.permissionOp.SqlxExecute(
		ctx,
		`insert into user_roles (created,user,role) values(?,?,?)`,
		time.Now().Unix(), uid, r.Id,
	)
}

func (backend *_BackendT) DelRoleForUser(ctx context.Context, role string, uid int64) {
	r := backend.roleOp.getByName(ctx, role)
	if r == nil {
		panic(fmt.Errorf("nil role, `%s`", role))
	}

	backend.permissionOp.SqlxExecute(
		ctx,
		`update user_roles set deleted=? where user=? and role=? and deleted=0`,
		time.Now().Unix(), uid, r.Id,
	)
}

var backend *_BackendT
var Rbac rbac.Rbac

func init() {
	internal.LazyExecutor.RegisterWithPriority(
		func(args snow.Kwargs) {
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(_RolePermissions{})))
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(_UserRoles{})))

			backend = &_BackendT{
				permissionOp: permissionOp,
				roleOp:       roleOp,
			}

			backend.permissionOp.Init(reflect.TypeOf(Permission{}))
			backend.roleOp.Init(reflect.TypeOf(Role{}))

			Rbac = rbac.Default(backend)
		},
		permissionPriority.Incr().Incr(), // after permission ensure
	)
}
