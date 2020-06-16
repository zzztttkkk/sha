package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"google.golang.org/grpc"
	"log"
	"net"
)

func RunAsGrpcServer(loader *_LoaderT, server *grpc.Server) {
	loader.bindGrpc(server)

	listener, err := net.Listen("tcp", string(ini.GetMust("services.grpc.address")))
	if err != nil {
		panic(err)
	}
	log.Fatal(server.Serve(listener))
}
