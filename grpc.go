package snow

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

func RunAsGrpcServer(loader *_LoaderT, server *grpc.Server) {
	loader.bindGrpc(server)

	listener, err := net.Listen("tcp", string(config.GetMust("services.grpc.address")))
	if err != nil {
		panic(err)
	}
	log.Fatal(server.Serve(listener))
}
