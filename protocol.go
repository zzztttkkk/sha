package suna

import (
	"context"
	"net"
)

type Protocol interface {
	Serve(ctx context.Context, s *Server, conn net.Conn)
	Upgrade(connCtx context.Context, conn net.Conn, ctx *RequestCtx, name []byte) Protocol
}
