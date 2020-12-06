package rbac

import (
	"context"
	sunainternal "github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/rbac/dao"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/sqlx"
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
		sunainternal.Silence(func() { dao.RoleAddPerm(ctx, root, perm.Name) })
	}

	// grant role
	sunainternal.Silence(func() { dao.GrantRole(ctx, root, subjectID) })
}

func GrantRoot(subjectID int64) {
	if inited {
		grantRoot(subjectID)
		return
	}
	internal.Dig.Append(func(_ _PermOK) { grantRoot(subjectID) })
}
