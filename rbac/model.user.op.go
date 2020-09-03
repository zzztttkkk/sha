package rbac

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"strconv"
)

type userOpT struct {
	roles *sqls.Operator
	lru   *utils.Lru
}

var _UserOperator = &userOpT{
	roles: &sqls.Operator{},
}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_UserOperator.roles.Init(userWithRoleT{})
			_UserOperator.lru = utils.NewLru(cfg.Cache.Lru.UserSize)
		},
		permTablePriority.Incr(),
	)
}

func (op *userOpT) changeRole(ctx context.Context, subjectId int64, roleName string, mt modifyType) error {
	roleId := _RoleOperator.GetIdByName(ctx, roleName)
	if roleId < 1 {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}
	defer LogOperator.Create(
		ctx,
		"user.changeRole",
		utils.M{"user": subjectId, "role": fmt.Sprintf("%d:%s", roleId, roleName), "modify": mt.String()},
	)
	defer op.lru.Remove(strconv.FormatInt(subjectId, 16))

	cond := sqls.AND().
		Eq(true, "role", roleId).
		Eq(true, "subject", subjectId)

	srb := op.roles.SelectBuilder(ctx, "role").From(op.roles.TableName()).Where(cond)

	var _id int64
	op.roles.ExecuteSelect(ctx, &_id, srb)
	if _id < 1 {
		if mt == _Add {
			return nil
		}

		kvs := utils.AcquireKvs()
		defer kvs.Free()
		kvs.Set("subject", subjectId)
		kvs.Set("role", roleId)

		op.roles.ExecuteCreate(ctx, kvs)
		return nil
	}

	if mt == _Add {
		return nil
	}

	q, args, err := op.roles.DeleteBuilder(ctx).From(op.roles.TableName()).Where(cond).Limit(1).ToSql()
	if err != nil {
		panic(err)
	}
	op.roles.ExecuteSql(ctx, q, args...)
	return nil
}

func (op *userOpT) getRoles(ctx context.Context, userId int64) []int64 {
	v, ok := op.lru.Get(strconv.FormatInt(userId, 16))
	if ok {
		return v.([]int64)
	}

	lst := make([]int64, 0)
	op.roles.ExecuteSelect(
		ctx,
		&lst,
		op.roles.SelectBuilder(ctx, "role").From(op.roles.TableName()).Prefix("distinct").Where(
			"subject=? and role>0 and status>=0 and deleted=0", userId,
		).OrderBy("role"),
	)
	op.lru.Add(strconv.FormatInt(userId, 16), lst)
	return lst
}
