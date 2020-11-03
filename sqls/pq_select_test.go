package sqls

import (
	"context"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"testing"
	"time"
)

type A struct {
	Model
}

type _AOp struct {
	Operator
}

var Op = &_AOp{}

func init() {
	conf := config.Default()
	conf.Sql.Driver = "postgres"
	conf.Sql.Leader = "postgres://postgres:123456@localhost/suna_examples_simple?sslmode=disable"
	conf.Sql.Logging = true

	conf.Done()

	internal.Dig.Provide(
		func() *config.Suna { return conf },
	)
	internal.Dig.Provide(func() auth.Authenticator { return auth.Authenticator(nil) })
}

func TestInsert(t *testing.T) {
	internal.Dig.Invoke()
	Op.Init(A{})

	ctx, committer := Tx(context.Background())
	Op.ExecInsert(ctx, Insert("created").Values(time.Now().Unix()))
	committer()
}
