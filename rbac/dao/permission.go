package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/rbac/model"
	"github.com/zzztttkkk/suna/sqlx"
	"time"
)

type _PermOk int

var permOp *sqlx.Operator

func init() {
	internal.Dig.Provide(
		func(_ _LogOk) _PermOk {
			permOp = sqlx.NewOperator(model.Permission{})
			permOp.CreateTable(true)
			return 0
		},
	)
}

func Perms(ctx context.Context) []*model.Permission {
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
	defer logging(ctx, "w.new.perm", sqlx.JsonObject{"name": name, "description": desc})

	permOp.Insert(
		ctx,
		sqlx.Data{
			"created_at":  time.Now().Unix(),
			"name":        name,
			"description": desc,
		},
	)
}

func EnsurePerm(name string) string {
	ctx, committer := sqlx.Tx(context.Background())
	defer committer()

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
			"description": fmt.Sprintf("created by `suna.rbac.dao.EnsurePerm`"),
		},
	)
	return name
}

func DelPerm(ctx context.Context, name string) {
	defer logging(ctx, "w.del.perm", sqlx.JsonObject{"name": name})

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
	defer logging(ctx, "r.idbyname.perm", sqlx.JsonObject{"name": name})

	type Arg struct {
		Name string `db:"name"`
	}

	var id int64
	err := permOp.RowColumns(
		ctx,
		"id",
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
