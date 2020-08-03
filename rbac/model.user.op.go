package rbac

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/cache"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"strconv"
)

type _UserOp struct {
	roles *sqls.Operator
	lru   *cache.Lru
}

var _UserOperator = &_UserOp{
	roles: &sqls.Operator{},
}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_UserOperator.roles.Init(reflect.ValueOf(_UserWithRole{}))
			cache.NewLru(cfg.Cache.Lru.UserSize)
		},
		permTablePriority.Incr(),
	)
}

func (op *_UserOp) changeRole(ctx context.Context, subjectId int64, roleName string, mt modifyType) error {
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

	q, vl := op.roles.BindNamed(
		fmt.Sprintf("select role from %s where role=:rid and user=:uid", op.roles.TableName()),
		utils.M{"rid": roleId, "uid": subjectId},
	)
	var _id int64
	op.roles.XQ11(ctx, &_id, q, vl...)
	if _id < 1 {
		if mt == Add {
			return nil
		}
		op.roles.XCreate(
			ctx,
			utils.M{
				"user": subjectId,
				"role": roleId,
			},
		)
		return nil
	}

	if mt == Add {
		return nil
	}

	q, vl = op.roles.BindNamed("delete from %s where role=:rid and user=:uid", utils.M{"rid": roleId, "uid": subjectId})
	op.roles.XExecute(ctx, q, vl...)
	return nil
}

func (op *_UserOp) getRoles(ctx context.Context, userId int64) []int64 {
	v, ok := op.lru.Get(strconv.FormatInt(userId, 16))
	if ok {
		return v.([]int64)
	}
	var lst []int64
	op.roles.XQ1n(ctx, &lst, fmt.Sprintf(`select distinct role from %s where user=? and role>0 order by role`, op.roles.TableName()), userId)
	op.lru.Add(strconv.FormatInt(userId, 16), lst)
	return lst
}
