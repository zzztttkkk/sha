package unregister

import (
	"context"
	"github.com/zzztttkkk/suna.example/model"
	"github.com/zzztttkkk/suna/sqls"
)

func Do(ctx context.Context, user *model.User, skey string) {
	ctx, committer := sqls.Tx(ctx)
	defer committer()
	model.UserOperator.Delete(ctx, user, skey)
}
