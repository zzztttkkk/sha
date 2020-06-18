package models

import (
	"context"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
)

type _PermStringT string

func (perm _PermStringT) ensure() _PermStringT {
	internal.LazyExecutor.RegisterWithPriority(
		func(kwargs snow.Kwargs) {
			ctx, committer := sqls.Tx(context.Background())
			defer committer()

			permissionOp.ensure(ctx, string(perm))
		},
		permissionPriority.Incr(), // after operator init
	)
	return perm
}

var PermPermissionCreate = _PermStringT("admin.perm.create").ensure()
var PermPermissionDelete = _PermStringT("admin.perm.delete").ensure()
var PermRoleCreate = _PermStringT("admin.role.create").ensure()
var PermRoleDelete = _PermStringT("admin.role.delete").ensure()
var PermRoleAddPermission = _PermStringT("admin.role.perm.add").ensure()
var PermRoleDelPermission = _PermStringT("admin.role.perm.del").ensure()
var PermRoleGrantToSubject = _PermStringT("admin.role.grant2subject").ensure()
var PermRoleRemoveForSubject = _PermStringT("admin.role.remove4subject").ensure()

var PermLogin = _PermStringT("user.account.login").ensure()
var PermCreatePost = _PermStringT("user.post.create").ensure()
