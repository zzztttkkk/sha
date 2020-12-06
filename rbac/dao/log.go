package dao

import (
	"context"
	sunainternal "github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/rbac/auth"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/rbac/model"
	"github.com/zzztttkkk/suna/sqlx"
	"time"
)

var logOp *sqlx.Operator

type _LogOk int

func init() {
	internal.Dig.Provide(
		func(_ internal.AuthOK) _LogOk {
			logOp = sqlx.NewOperator(model.Log{})
			logOp.CreateTable(true)
			return 0
		},
	)
}

var LogReadOperation bool

type _LogItem struct {
	name      string
	subjectID int64
	info      sqlx.JsonObject
}

var ch = make(chan *_LogItem, 256)

func init() {
	go func() {
		sleepStem := time.Microsecond * 10

		for {
			select {
			case v := <-ch:
				sunainternal.Silence(
					func() {
						ctx, committer := sqlx.Tx(context.Background())

						logOp.Insert(
							ctx,
							sqlx.Data{
								"created_at": time.Now().Unix(),
								"name":       v.name,
								"operator":   v.subjectID,
								"info":       v.info,
							},
						)

						committer()
					},
				)
				sleepStem = time.Microsecond * 10
			default:
				time.Sleep(sleepStem)
				sleepStem *= 2
				if sleepStem > time.Second {
					sleepStem = time.Second
				}
			}
		}
	}()
}

func logging(ctx context.Context, name string, info sqlx.JsonObject) {
	if name[0] == 'r' && !LogReadOperation {
		return
	}

	ch <- &_LogItem{
		name:      name,
		subjectID: auth.MustAuth(ctx).GetID(),
		info:      info,
	}
}
