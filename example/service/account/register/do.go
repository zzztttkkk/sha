package register

import (
	"context"
	"github.com/zzztttkkk/suna.example/model"
	"github.com/zzztttkkk/suna/sqls"
)

func Do(ctx context.Context, name, password []byte) (int64, []byte) {
	ctx, committer := sqls.Tx(ctx)
	defer committer()
	return model.UserOperator.Create(ctx, name, password)
}
