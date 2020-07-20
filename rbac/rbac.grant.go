package rbac

import "context"

func IsGranted(ctx context.Context, subject Subject, permission string) bool {
	g.RLock()
	defer g.RUnlock()

	SubjectId := subject.GetId()
	if SubjectId < 1 {
		return false
	}

	Perm, ok := permNameMap[permission]
	if !ok {
		return false
	}

	Pid := Perm.GetId()

	// private subject can has permissions
	if SubjectId < MaxPrivateSubjectId {
		for _, pid := range backend.GetSubjectPermissions(ctx, subject.GetId()) {
			if pid == Pid {
				return true
			}
		}
	}

	for _, rid := range backend.GetSubjectRoles(ctx, subject.GetId()) {
		permMap, ok := rolePermMap[rid]
		if ok {
			if permMap[Pid] {
				return true
			}
		}
	}
	return false
}
