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
	// read
	GetAllPermissions(ctx context.Context) []Permission
	GetAllRoles(ctx context.Context) []Role

	GetSubjectRoles(ctx context.Context, subjectId int64) []int64
	GetSubjectPermissions(ctx context.Context, subjectId int64) []int64

	// write
	NewPermission(ctx context.Context, name string) error
	NewRole(ctx context.Context, name string) error
	RoleAddPermission(ctx context.Context, roleName, permName string) error
	RoleDelPermission(ctx context.Context, roleName, permName string) error
	SubjectAddRole(ctx context.Context, subject int64, roleName string) error
	SubjectDelRole(ctx context.Context, subject int64, roleName string) error
	SubjectAddPermission(ctx context.Context, subject int64, permissionName string) error
	SubjectDelPermission(ctx context.Context, subject int64, permissionName string) error
}
