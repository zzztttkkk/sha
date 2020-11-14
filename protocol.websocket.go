package suna

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"net"
	"net/http"
	"sync"
)

type WebsocketProtocol struct {
	server   *Server
	PreCheck func(ctx *RequestCtx) bool
}

var headerWebSocketSecretKey = []byte("Sec-WebSocket-Key")
var headerWebSocketVersion = []byte("Sec-WebSocket-Version")
var headerWebSocketAccept = []byte("Sec-WebSocket-Accept")
var protocolSecret = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
var websocketStr = []byte("websocket")

type _WebsocketHash struct {
	hash.Hash
	data []byte
}

var hashPool = sync.Pool{New: func() interface{} { return &_WebsocketHash{Hash: sha1.New(), data: make([]byte, 28)} }}

func (protocol *WebsocketProtocol) Handshake(ctx *RequestCtx) bool {
	secret, ok := ctx.Request.Header.Get(headerWebSocketSecretKey)
	if !ok || len(secret) < 1 {
		ctx.WriteError(StdHttpErrors[http.StatusBadRequest])
		return false
	}
	version, ok := ctx.Request.Header.Get(headerWebSocketVersion)
	if !ok || len(version) != 2 || version[0] != '1' || version[1] != '3' {
		ctx.WriteError(StdHttpErrors[http.StatusBadRequest])
		return false
	}

	if protocol.PreCheck != nil {
		if !protocol.PreCheck(ctx) {
			if ctx.Response.statusCode == 0 {
				ctx.WriteError(StdHttpErrors[http.StatusBadRequest])
			}
			return false
		}
	}

	h := hashPool.Get().(*_WebsocketHash)
	_, _ = h.Write(secret)
	_, _ = h.Write(protocolSecret)
	base64.StdEncoding.Encode(h.data, h.Sum(nil))

	ctx.WriteError(StdHttpErrors[http.StatusSwitchingProtocols])
	ctx.Response.Header.Set(upgradeHeader, websocketStr)
	ctx.Response.Header.Set(headerWebSocketAccept, h.data)
	return true
}

func (protocol *WebsocketProtocol) Serve(ctx context.Context, conn net.Conn, request *Request) {
	fmt.Println(conn.RemoteAddr().String(), string(request.Path))
	// todo
}
