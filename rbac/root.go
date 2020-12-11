package rbac

import (
	"context"
	shainternal "github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
)

const root = "rbac.root"

func grantRoot(subjectID int64) {
	ctx, committer := sqlx.Tx(context.Background())
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
	shainternal.Silence(func() { dao.GrantRole(ctx, root, subjectID) })
}

func GrantRoot(subjectID int64) {
	if shainternal.RbacInited {
		grantRoot(subjectID)
		return
	}
	internal.Dig.Append(func(_ _PermOK) { grantRoot(subjectID) })
}
