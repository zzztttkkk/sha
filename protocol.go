package suna

import (
	"context"
	"net"
)

type Protocol interface {
	Serve(ctx context.Context, s *Server, conn net.Conn)
}


