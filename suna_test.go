package suna

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/utils"
	"testing"
)

var REDIS_URL = "redis://127.0.0.1:6379"
var SQL_URL = ":memory:"
var SQL_DRIVER = "sqlite3"

func TestInit(t *testing.T) {
	conf := config.Default()
	conf.Redis.Mode = "singleton"
	conf.Redis.Nodes = append(conf.Redis.Nodes, REDIS_URL)
	conf.Sql.Leader = SQL_URL
	conf.Sql.Driver = SQL_DRIVER
	conf.Sql.Logging = true

	conf.Done()

	Init(&InitOption{Config: &conf})

	root := router.New()
	root.GET(
		"/hello",
		func(ctx *fasthttp.RequestCtx) { output.MsgOK(ctx, "World!") },
	)

	loader := utils.NewLoader()
	loader.AddChild("/rbac", rbac.Loader())

	loader.RunAsHttpServer(root, &conf)
}
