package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/ctxs"
	"github.com/zzztttkkk/snow/examples/blog/backend"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/examples/blog/backend/services"
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/utils"
	"log"
	"time"
)

func main() {
	conf := &snow.Config{}
	conf.IniFiles = append(conf.IniFiles, "examples/blog/conf.ini")
	conf.UserReader = models.UserReader
	snow.Init(conf)

	backend.Init()

	root := router.New()
	root.Use(
		mware.NewRateLimiter(ctxs.GetRemoteIpHash, time.Second, 30),
		mware.SessionHandler,
	)

	services.Loader.BindHttp(root)

	rlog := utils.AcquireGroupLogger("Router")
	for method, paths := range root.List() {
		for _, path := range paths {
			rlog.Println(fmt.Sprintf("%s: %s", method, path))
		}
	}
	utils.ReleaseGroupLogger(rlog)

	log.Fatal(fasthttp.ListenAndServe(ini.MustGet("services.http.address"), root.Handler))
}
