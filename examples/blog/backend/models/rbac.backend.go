package models

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/rbac"
	"github.com/zzztttkkk/snow/sqls"
)

type _BackendT struct {
	permissionOp *_PermissionOperator
	roleOp       *_RoleOperator
}

func (backend *_BackendT) Flush(ctx context.Context) {
	RbacInstance.Reload(ctx)
}

func (backend *_BackendT) GetAllPermissions(ctx context.Context) []rbac.Permission {
	return backend.permissionOp.getAll(ctx)
}

func (backend *_BackendT) GetAllRoles(ctx context.Context) []rbac.Role {
	return backend.roleOp.getAll(ctx)
}

func (backend *_BackendT) AddPermission(ctx context.Context, permission, descp string) {
	backend.permissionOp.createIfNotExists(ctx, permission, descp)
}

func (backend *_BackendT) DelPermission(ctx context.Context, permission string) {
	backend.permissionOp.XUpdate(
		ctx,
		`update permission set deleted=?,name=? where name=?`,
		time.Now().Unix(), fmt.Sprintf("D<%s>", permission), permission,
	)
}

func (backend *_BackendT) AddRole(ctx context.Context, role, descp string) {
	backend.roleOp.XCreate(
		ctx,
		sqls.Dict{"name": role, "created": time.Now().Unix(), "descp": descp},
	)
}

func (backend *_BackendT) DelRole(ctx context.Context, role string) {
	backend.permissionOp.XUpdate(
		ctx,
		`update role set deleted=?,name=? where name=?`,
		time.Now().Unix(), fmt.Sprintf("D<%s>", role), role,
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

	backend.permissionOp.XExecute(
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

	backend.permissionOp.XExecute(
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

	backend.permissionOp.XExecute(
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

	backend.permissionOp.XExecute(
		ctx,
		`update user_roles set deleted=? where user=? and role=? and deleted=0`,
		time.Now().Unix(), uid, r.Id,
	)
}

var RbacBackend *_BackendT
var RbacInstance rbac.Rbac

func init() {
	internal.LazyExecutor.RegisterWithPriority(
		func(args snow.Kwargs) {
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(_RolePermissions{})))
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(_UserRoles{})))

			RbacBackend = &_BackendT{
				permissionOp: permissionOp,
				roleOp:       roleOp,
			}

			RbacInstance = rbac.Default(RbacBackend)
		},
		rbacInitPriority,
	)
}
