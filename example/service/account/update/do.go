package update

import (
	"context"
	"github.com/zzztttkkk/suna.example/model"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
)

func Do(ctx context.Context, uid int64, secret []byte, kvs *utils.Kvs) bool {
	if kvs.Len() < 1 {
		return true
	}

	ctx, committer := sqls.Tx(ctx)
	defer committer()

	return model.UserOperator.Update(ctx, uid, kvs, secret)
}
