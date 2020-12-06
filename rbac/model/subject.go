package model

import x "github.com/jmoiron/sqlx"

type SubjectWithRoles struct {
	Subject int64
	Role    int64
}

func (swr SubjectWithRoles) TableName() string {
	return TablenamePrefix + "subject_with_roles"
}

func (swr SubjectWithRoles) TableColumns(_ *x.DB) []string {
	return []string{
		"subject bigint not null",
		"role bigint not null",
		"primary key(role,subject)",
	}
}
