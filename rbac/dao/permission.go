package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/rbac/model"
	"github.com/zzztttkkk/sha/sqlx"
	"time"
)

type _PermOk int

var permOp *sqlx.Operator

func init() {
	internal.Dig.Provide(
		func() _PermOk {
			permOp = sqlx.NewOperator(model.Permission{})
			permOp.CreateTable()
			return 0
		},
	)
}

func Perms(ctx context.Context) []*model.Permission {
	logging(ctx, "r.perms", nil)

	var ret []*model.Permission
	err := permOp.FetchMany(ctx, "*", "where deleted_at=0 and status>=0", nil, &ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return ret
		}
		panic(err)
	}
	return ret
}

func NewPerm(ctx context.Context, name, desc string) {
	logging(ctx, "w.new.perm", sqlx.JsonObject{"name": name, "description": desc})

	permOp.Insert(
		ctx,
		sqlx.Data{
			"created_at":  time.Now().Unix(),
			"name":        name,
			"description": desc,
		},
	)
}

func CreatePermIfNotExists(name string) string {
	ctx, tx := sqlx.Tx(context.Background())
	defer tx.AutoCommit(ctx)

	_, e := GetPermIDByName(ctx, name)
	if e != nil {
		if e != sql.ErrNoRows {
			panic(e)
		}
	} else {
		return name
	}

	permOp.Insert(
		ctx,
		sqlx.Data{
			"created_at":  time.Now().Unix(),
			"name":        name,
			"description": fmt.Sprintf("created by `sha.rbac.dao.EnsurePerm`"),
		},
	)
	return name
}

func DelPerm(ctx context.Context, name string) {
	logging(ctx, "w.del.perm", sqlx.JsonObject{"name": name})

	pid, err := GetPermIDByName(ctx, name)
	if err != nil {
		panic(err)
	}

	type Arg1 struct {
		Name string `db:"name"`
	}

	type Arg2 struct {
		PID int64 `db:"pid"`
	}

	roleWithPermOp.Delete(ctx, "where perm=:pid", Arg2{PID: pid})

	permOp.Update(
		ctx,
		sqlx.Data{"deleted_at": time.Now().Unix()},
		"where name=:name and deleted_at=0",
		Arg1{Name: name},
	)
}

func GetPermIDByName(ctx context.Context, name string) (int64, error) {
	logging(ctx, "r.idbyname.perm", sqlx.JsonObject{"name": name})

	type Arg struct {
		Name string `db:"name"`
	}

	var id int64
	err := permOp.RowColumns(
		ctx,
		[]string{"id"},
		"where name=:name and deleted_at=0 and status>=0",
		Arg{Name: name},
		&id,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func MustGetPermIDByName(ctx context.Context, name string) int64 {
	v, e := GetPermIDByName(ctx, name)
	if e != nil {
		panic(e)
	}
	return v
}
