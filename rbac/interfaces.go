package rbac

import (
	"context"
)

type Permission interface {
	GetId() int64
	GetName() string
}

type Role interface {
	GetId() int64
	GetName() string
	GetParentId() int64
	GetPermissionIds() []int64
}

type Subject interface {
	GetId() int64
}

type Backend interface {
	GetAllPermissions(ctx context.Context) []Permission
	GetAllRoles(ctx context.Context) []Role

	GetSubjectRoles(ctx context.Context, subjectId int64) []int64
	GetSubjectPermissions(ctx context.Context, subjectId int64) []int64
}
