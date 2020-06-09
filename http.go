package snow

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/utils"
	"log"
)

func RunAsHttpServer(loader *_LoaderT, root *router.Router) {
	loader.bindHttp(root)

	glog := utils.AcquireGroupLogger("Router")
	for method, paths := range root.List() {
		for _, path := range paths {
			glog.Println(fmt.Sprintf("%s: %s", method, path))
		}
	}
	utils.ReleaseGroupLogger(glog)
	log.Fatal(fasthttp.ListenAndServe(string(ini.GetMust("services.http.address")), root.Handler))
}
