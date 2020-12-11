package rbac

import (
	"context"
	"github.com/gobuffalo/packr/v2"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"html/template"
)

type ReqWriter interface {
	MustValidate(dist interface{})
	SetStatus(v int)
	Write(p []byte) (int, error)
	WriteHTML(f []byte)
	WriteJSON(v interface{})
	WriteTemplate(t *template.Template, data interface{})
}

type HandlerFunc func(rctx context.Context, rw ReqWriter)

type Router interface {
	HandleWithDoc(
		method string, path string,
		handler HandlerFunc,
		doc interface{},
	)
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

var box *packr.Box

func init() {
	internal.Dig.Provide(
		func(router Router, _ internal.DaoOK) _PermOK {
			dao.EnsurePerm(PermPermissionCreate)
			dao.EnsurePerm(PermPermissionDelete)
			dao.EnsurePerm(PermPermissionListAll)

			dao.EnsurePerm(PermRoleCreate)
			dao.EnsurePerm(PermRoleDelete)
			dao.EnsurePerm(PermRoleListAll)
			dao.EnsurePerm(PermRoleAddPerm)
			dao.EnsurePerm(PermRoleDelPerm)
			dao.EnsurePerm(PermRoleAddBased)
			dao.EnsurePerm(PermRoleDelBased)

			dao.EnsurePerm(PermRbacLogin)
			dao.EnsurePerm(PermGrantRole)
			dao.EnsurePerm(PermCancelRole)

			box = packr.New("", "./html")

			return 0
		},
	)
}
