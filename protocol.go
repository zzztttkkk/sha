package suna

import (
	"context"
	"net"
)

type Protocol interface {
	Handshake(ctx *RequestCtx) bool
	Serve(ctx context.Context, req *Request, conn net.Conn)
}
