package rbac

import (
	"context"
	"fmt"
	"strconv"

	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
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
	OP := op.roles

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

	cond := sqls.STR("role=? and subject=?", roleId, subjectId)
	var _id int64
	OP.ExecSelect(ctx, &_id, sqls.Select("role").Where(cond))
	if _id < 1 {
		if mt == _Add {
			return nil
		}
		OP.ExecInsert(ctx, sqls.Insert("subject, role").Values(subjectId, roleId))
		return nil
	}

	if mt == _Add {
		return nil
	}

	OP.ExecDelete(ctx, sqls.Delete().Where(cond).Limit(1))
	return nil
}

func (op *userOpT) getRoles(ctx context.Context, userId int64) []int64 {
	v, ok := op.lru.Get(strconv.FormatInt(userId, 16))
	if ok {
		return v.([]int64)
	}

	OP := op.roles
	lst := make([]int64, 0)
	OP.ExecSelect(
		ctx,
		&lst,
		sqls.Select("role").Distinct().
			Where("subject=? and role>0 and status>=0 and deleted=0", userId).OrderBy("role"),
	)
	op.lru.Add(strconv.FormatInt(userId, 16), lst)
	return lst
}
