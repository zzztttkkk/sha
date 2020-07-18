package snow

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/router"
	"github.com/zzztttkkk/snow/utils"
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
	log.Fatal(fasthttp.ListenAndServe(string(config.GetMust("services.http.address")), root.Handler))
}
