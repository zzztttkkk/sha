package rbac

import (
	"context"
)

func Load(ctx context.Context) {
	g.Lock()
	defer g.Unlock()

	permIdMap = map[int64]Permission{}
	permNameMap = map[string]Permission{}
	roleIdMap = map[int64]Role{}
	rolePermMap = map[int64]map[int64]bool{}

	for _, p := range backend.GetAllPermissions(ctx) {
		permIdMap[p.GetId()] = p
		permNameMap[p.GetName()] = p
	}

	for _, r := range backend.GetAllRoles(ctx) {
		roleIdMap[r.GetId()] = r
	}

	buildRolePermMap()
}

func buildRolePermMap() {
	for rid, role := range roleIdMap {
		rolePermMap[rid] = makeOneRole(role)
	}
}

func makeOneRole(role Role) map[int64]bool {
	pm := map[int64]bool{}
	err := false

	_makeOneRole(role, map[int64]bool{}, pm, &err)

	if err {
		return nil
	}
	return pm
}

func _makeOneRole(role Role, fp map[int64]bool, pm map[int64]bool, ep *bool) {
	_, ok := fp[role.GetId()]
	if ok {
		*ep = true
		return
	}

	for _, pid := range role.GetPermissionIds() {
		pm[pid] = true
	}

	rpid := role.GetParentId()
	if rpid > 0 {
		rp := roleIdMap[rpid]
		if rp == nil {
			*ep = true
			return
		}
		_makeOneRole(rp, fp, pm, ep)
	}
}
