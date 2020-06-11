package rbac

import "fmt"

type Error struct {
	SubjectId      int64
	PermissionName string
}

func (err *Error) Error() string {
	return fmt.Sprintf("snow.rbac: `%d` has no permission `%s`", err.SubjectId, err.PermissionName)
}
