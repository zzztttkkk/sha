package dao

import (
	"context"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/sqlx"
)

var LogReadOperation bool

func logging(ctx context.Context, name string, info sqlx.JsonObject) {
	if name[0] == 'r' && !LogReadOperation {
		return
	}
	subject := auth.MustAuth(ctx)
	internal.Logger.Printf("(%s) (%d %v) %v\n", name, subject.GetID(), subject.Info(), info)
}
