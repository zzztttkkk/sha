package dao

import (
	"context"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/sqlx"
)

var LogReadOperation bool

func logging(ctx context.Context, name string, info sqlx.JsonObject) {
	if name[0] == 'r' && !LogReadOperation {
		return
	}

	var id int64
	var sinfo interface{}

	if v := ctx.Value(internal.RootCtxKey); v != nil {
		rootCtx := v.(*internal.RootCtx)
		id = rootCtx.Uid
		sinfo = rootCtx.Info
	} else {
		subject := auth.MustAuth(ctx)
		id = subject.GetID()
		sinfo = subject.Info(ctx)
	}

	internal.Logger.Printf("(%s) (%d %v) %v\n", name, id, sinfo, info)
}
