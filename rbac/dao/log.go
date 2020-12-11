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
	subject := auth.MustAuth(ctx)
	internal.Logger.Printf("(%s) (%d %v) %v\n", name, subject.GetID(), subject.Info(), info)
}
