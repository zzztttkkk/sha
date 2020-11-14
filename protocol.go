package suna

import (
	"context"
	"net"
)

type Protocol interface {
	Handshake(ctx *RequestCtx) bool
	Serve(ctx context.Context, conn net.Conn, request *Request)
}
