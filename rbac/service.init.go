package rbac

import (
	"context"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/validator"
	"html/template"
)

type RCtx interface {
	context.Context
	MustValidate(dist interface{})
	SetStatus(v int)
	Write(p []byte) (int, error)
	WriteHTML(f []byte)
	WriteJSON(v interface{})
	WriteTemplate(t *template.Template, data interface{})
}

type HandlerFunc func(rctx RCtx)

type Router interface {
	HandleWithDoc(method string, path string, handler HandlerFunc, doc validator.Document)
}

const (
	PermPermissionCreate  = "rbac.permission.create"
	PermPermissionDelete  = "rbac.permission.delete"
	PermPermissionListAll = "rbac.permission.list"

	PermRoleCreate   = "rbac.role.create"
	PermRoleDelete   = "rbac.role.delete"
	PermRoleListAll  = "rbac.role.list"
	PermRoleAddPerm  = "rbac.role.add_perm"
	PermRoleDelPerm  = "rbac.role.del_perm"
	PermRoleAddBased = "rbac.role.add_based"
	PermRoleDelBased = "rbac.role.del_based"

	PermRbacLogin  = "rbac.login"
	PermGrantRole  = "rbac.grant_role"
	PermCancelRole = "rbac.cancel_role"
)

type _PermOK int

func init() {
	internal.Dig.Provide(
		func(router Router, _ internal.DaoOK) _PermOK {
			dao.CreatePermIfNotExists(PermPermissionCreate)
			dao.CreatePermIfNotExists(PermPermissionDelete)
			dao.CreatePermIfNotExists(PermPermissionListAll)

			dao.CreatePermIfNotExists(PermRoleCreate)
			dao.CreatePermIfNotExists(PermRoleDelete)
			dao.CreatePermIfNotExists(PermRoleListAll)
			dao.CreatePermIfNotExists(PermRoleAddPerm)
			dao.CreatePermIfNotExists(PermRoleDelPerm)
			dao.CreatePermIfNotExists(PermRoleAddBased)
			dao.CreatePermIfNotExists(PermRoleDelBased)

			dao.CreatePermIfNotExists(PermRbacLogin)
			dao.CreatePermIfNotExists(PermGrantRole)
			dao.CreatePermIfNotExists(PermCancelRole)

			return 0
		},
	)
}
