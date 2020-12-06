package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zzztttkkk/suna/auth"
	sunainternal "github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/rbac/model"
	"github.com/zzztttkkk/suna/sqlx"
	"net/http"
	"time"
)

var roleOp *sqlx.Operator
var roleWithPermOp *sqlx.Operator
var roleWithInhOp *sqlx.Operator
var subjectWithRoleOp *sqlx.Operator

// sqls
var rolePermsSql = "select * from %s as p where p.id in (select rp.perm from %s as rp where rp.role=:rid) order by p.id"

func init() {
	internal.Dig.Provide(
		func(_ _PermOk) internal.DaoOK {
			roleOp = sqlx.NewOperator(model.Role{})
			roleOp.CreateTable(true)

			roleWithPermOp = sqlx.NewOperator(model.RoleWithPermissions{})
			roleWithPermOp.CreateTable(true)

			roleWithInhOp = sqlx.NewOperator(model.RoleWithInheritances{})
			roleWithInhOp.CreateTable(true)

			subjectWithRoleOp = sqlx.NewOperator(model.SubjectWithRoles{})
			subjectWithRoleOp.CreateTable(true)

			rolePermsSql = fmt.Sprintf(rolePermsSql, permOp.TableName(), roleWithPermOp.TableName())
			return 0
		},
	)
}

func Roles(ctx context.Context) []*model.Role {
	var ret []*model.Role
	err := roleOp.FetchMany(ctx, "*", "where deleted_at=0 and status>=0", nil, &ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return ret
		}
		panic(err)
	}
	return ret
}

func RoleByName(ctx context.Context, name string) *model.Role {
	var ret model.Role

	type Arg struct {
		Name string `db:"name"`
	}
	err := roleOp.FetchOne(ctx, "*", "where deleted_at=0 and status>=0 and name=:rname", Arg{Name: name}, &ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &ret
}

func NewRole(ctx context.Context, name, desc string) {
	defer logging(ctx, "w.new.role", sqlx.JsonObject{"name": name, "description": desc})

	roleOp.Insert(
		ctx,
		sqlx.Data{
			"created_at":  time.Now().Unix(),
			"name":        name,
			"description": desc,
		},
	)
}

func GetRoleIDByName(ctx context.Context, name string) (int64, error) {
	defer logging(ctx, "r.idbyname.role", sqlx.JsonObject{"name": name})

	type Arg struct {
		Name string `db:"name"`
	}

	var id int64
	err := roleOp.RowColumns(
		ctx, "id",
		"where name=:name and deleted_at=0 and status>=0 limit 1",
		Arg{Name: name},
		&id,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func MustGetRoleIDByName(ctx context.Context, name string) int64 {
	v, e := GetRoleIDByName(ctx, name)
	if e != nil {
		panic(e)
	}
	return v
}

func DelRole(ctx context.Context, name string) {
	rid := MustGetRoleIDByName(ctx, name)
	defer logging(ctx, "w.del.role", sqlx.JsonObject{"name": name})

	type Arg struct {
		RoleID int64 `db:"rid"`
	}

	d := Arg{RoleID: rid}

	roleWithInhOp.Delete(ctx, "where role=:rid", d)
	roleWithPermOp.Delete(ctx, "where role=:rid", d)
	subjectWithRoleOp.Delete(ctx, "where role=:rid", d)

	roleOp.Update(
		ctx,
		sqlx.Data{"deleted_at": time.Now().Unix()},
		"where id=:rid and deleted_at=0",
		d,
	)
}

func RoleAddPerm(ctx context.Context, role, perm string) {
	rid := MustGetRoleIDByName(ctx, role)
	pid := MustGetPermIDByName(ctx, perm)
	defer logging(ctx, "w.add.role.perm", sqlx.JsonObject{"role": role, "perm": perm})

	roleWithPermOp.Insert(ctx, sqlx.Data{"role": rid, "perm": pid})
}

func RoleDelPerm(ctx context.Context, role, perm string) {
	rid := MustGetRoleIDByName(ctx, role)
	pid := MustGetPermIDByName(ctx, perm)
	defer logging(ctx, "w.del.role.perm", sqlx.JsonObject{"role": role, "roleID": rid, "perm": perm, "permID": pid})

	type Arg struct {
		RID int64 `db:"rid"`
		PID int64 `db:"pid"`
	}

	roleWithPermOp.Delete(ctx, "where role=:rid and perm=:pid", Arg{RID: rid, PID: pid})
}

func RolePermIDs(ctx context.Context, roleID int64) []int64 {
	var ret []int64
	type Arg struct {
		RoleID int64 `db:"rid"`
	}
	err := roleWithPermOp.RowsColumn(ctx, "perm", "where role=:rid", Arg{RoleID: roleID}, &ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return ret
		}
		panic(err)
	}
	return ret
}

func RolePerms(ctx context.Context, roleID int64) []*model.Permission {
	var ret []*model.Permission

	type Arg struct {
		RID int64 `db:"rid"`
	}
	err := sqlx.Exe(ctx).RowsStruct(ctx, rolePermsSql, Arg{RID: roleID}, &ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return ret
		}
		panic(err)
	}
	return ret
}

func RoleGetBasedByID(ctx context.Context, roleID int64) []int64 {
	defer logging(ctx, "r.role.based", sqlx.JsonObject{"roleID": roleID})
	var ret []int64

	type Arg struct {
		RID int64 `db:"rid"`
	}

	_ = roleWithInhOp.RowsColumn(ctx, "based", "where role=:rid", Arg{RID: roleID}, &ret)
	return ret
}

func upTraverse(ctx context.Context, rid int64, footprint map[int64]struct{}) bool {
	_, exists := footprint[rid]
	if exists {
		return false
	}

	footprint[rid] = struct{}{}

	for _, v := range RoleGetBasedByID(ctx, rid) {
		if !upTraverse(ctx, v, footprint) {
			return false
		}
	}
	return true
}

func GetAllBasedRoleIDs(ctx context.Context, roleID int64) map[int64]struct{} {
	ret := map[int64]struct{}{}
	if !upTraverse(ctx, roleID, ret) {
		panic(ErrCircularReference)
	}
	return ret
}

var ErrCircularReference = errors.New("suna.rbac: circular reference")

func init() {
	sunainternal.ErrorStatusByValue[ErrCircularReference] = http.StatusInternalServerError
}

func RoleInheritFrom(ctx context.Context, role, base string) error {
	rid := MustGetRoleIDByName(ctx, role)
	bid := MustGetRoleIDByName(ctx, base)
	defer logging(ctx, "w.add.role.based", sqlx.JsonObject{"role": role, "based": base, "roleID": rid, "basedID": bid})

	ret := map[int64]struct{}{}
	if !upTraverse(ctx, rid, ret) {
		panic(ErrCircularReference)
	}
	if _, v := ret[bid]; v {
		return ErrCircularReference
	}
	roleWithInhOp.Insert(ctx, sqlx.Data{"role": rid, "based": bid})
	return nil
}

func RoleUninheritFrom(ctx context.Context, role, base string) {
	rid := MustGetRoleIDByName(ctx, role)
	bid := MustGetRoleIDByName(ctx, base)
	defer logging(ctx, "w.del.role.based", sqlx.JsonObject{"role": role, "based": base, "roleID": rid, "basedID": bid})

	type Arg struct {
		RID int64 `db:"rid"`
		BID int64 `db:"bid"`
	}

	roleWithInhOp.Delete(ctx, "where role=:rid and based=:bid", Arg{RID: rid, BID: bid})
}

func GrantRole(ctx context.Context, role string, subID int64) {
	rid := MustGetRoleIDByName(ctx, role)
	defer logging(ctx, "w.grant.subject.role", sqlx.JsonObject{"role": role, "roleID": rid, "subjectID": subID})

	subjectWithRoleOp.Insert(ctx, sqlx.Data{"subject": subID, "role": rid})
}

func CancelRole(ctx context.Context, role string, subID int64) {
	rid := MustGetRoleIDByName(ctx, role)
	defer logging(ctx, "w.cancel.subject.role", sqlx.JsonObject{"role": role, "roleID": rid, "subjectID": subID})

	type Arg struct {
		RID int64 `db:"rid"`
		SID int64 `db:"sid"`
	}

	subjectWithRoleOp.Delete(ctx, "where subject=:sid and role=:rid", Arg{RID: rid, SID: subID})
}

func SubjectRoles(ctx context.Context, subject auth.Subject) []int64 {
	defer logging(ctx, "r.subject.roles", sqlx.JsonObject{"subjectID": subject.GetID()})

	type Arg struct {
		SID int64 `db:"sid"`
	}

	var ret []int64
	err := subjectWithRoleOp.RowsColumn(ctx, "role", "where subject=:sid order by role", Arg{SID: subject.GetID()}, &ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return ret
		}
		panic(err)
	}
	return ret
}
