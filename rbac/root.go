package rbac

import (
	shainternal "github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
)

const root = "rbac.root"

func grantRoot(rootId int64, rootInfo interface{}) {
	ctx, committer := sqlx.Tx(internal.NewRootContext(rootId, rootInfo))
	defer committer()

	// create root role
	rootRoleId, _ := dao.GetRoleIDByName(ctx, root)
	if rootRoleId < 1 {
		dao.NewRole(ctx, root, "rbac root")
		rootRoleId, _ = dao.GetRoleIDByName(ctx, root)
	}

	// grant all rbac permissions
	for _, perm := range dao.Perms(ctx) {
		shainternal.Silence(func() { dao.RoleAddPerm(ctx, root, perm.Name) })
	}

	// grant role
	shainternal.Silence(func() { dao.GrantRole(ctx, root, rootId) })
}

func GrantRoot(rootId int64, rootInfo interface{}) {
	if shainternal.RbacInited {
		grantRoot(rootId, rootInfo)
		return
	}
	internal.Dig.Append(func(_ _PermOK) { grantRoot(rootId, rootInfo) })
}
