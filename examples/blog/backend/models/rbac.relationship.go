package models

type _RolePermissions struct {
	Id         int64 `ddl:"notnull;primary;incr"`
	Permission int64 `ddl:"notnull"`
	Role       int64 `ddl:"notnull"`
	Created    int64
	Deleted    int64 `ddl:"D<0>"`
}

func (_RolePermissions) TableName() string {
	return "role_permissions"
}

type _UserRoles struct {
	Id      int64 `ddl:"notnull;primary;incr"`
	User    int64 `ddl:"notnull"`
	Role    int64 `ddl:"notnull"`
	Created int64 `ddl:"notnull"`
	Deleted int64 `ddl:"D<0>"`
}

func (_UserRoles) TableName() string {
	return "user_roles"
}

