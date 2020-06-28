package models

import (
	"context"

	"github.com/zzztttkkk/snow/sqls"

	"github.com/zzztttkkk/snow"

	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
)

func NewPermission(name string) string {
	internal.LazyExecutor.RegisterWithPriority(
		func(kwargs snow.Kwargs) {
			ctx, committer := sqls.Tx(context.Background())
			defer committer()

			permissionOp.createIfNotExists(ctx, name, "")
		},
		permissionCreatePriority,
	)
	return name
}

var PermPermissionCreate = NewPermission("admin.perm.create")
var PermPermissionDelete = NewPermission("admin.perm.delete")
var PermRoleCreate = NewPermission("admin.role.create")
var PermRoleDelete = NewPermission("admin.role.delete")
var PermRoleAddPermission = NewPermission("admin.role.perm.add")
var PermRoleDelPermission = NewPermission("admin.role.perm.del")
var PermRoleGrantToSubject = NewPermission("admin.role.grant2subject")
var PermRoleRemoveForSubject = NewPermission("admin.role.remove4subject")

var PermLogin = NewPermission("user.account.login")
var PermCreatePost = NewPermission("user.post.create")
