package sha

import (
	"bytes"
	"context"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/websocket"
	"log"
	"net/http"
	"sync"
)

type WebSocketProtocol struct {
	ReadBufferSize    int
	WriteBufferSize   int
	Subprotocols      map[string]struct{}
	EnableCompression bool
}

var websocketStr = []byte("websocket")

const (
	websocketExtCompress = "permessage-deflate"
	websocketExt         = "permessage-deflate; server_no_context_takeover; client_no_context_takeover"
)

func (p *WebSocketProtocol) Handshake(ctx *RequestCtx) bool {
	version, _ := ctx.Request.Header.Get(HeaderSecWebSocketVersion)
	if len(version) != 2 || version[0] != '1' || version[1] != '3' {
		ctx.Response.statusCode = http.StatusBadRequest
		return false
	}
	if _, ok := ctx.Response.Header.Get(HeaderSecWebSocketExtensions); ok {
		log.Println("websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported")
		ctx.Response.statusCode = http.StatusInternalServerError
		return false
	}
	key, _ := ctx.Request.Header.Get(HeaderSecWebSocketKey)
	if len(key) < 1 {
		ctx.Response.statusCode = http.StatusBadRequest
		return false
	}

	var subprotocol []byte
	if len(p.Subprotocols) > 0 {
		hv, ok := ctx.Response.Header.Get(HeaderSecWebSocketProtocol)
		if ok {
			if _, ok := p.Subprotocols[utils.S(hv)]; ok {
				subprotocol = hv
				ctx.Request.webSocketSubProtocolName = append(ctx.Request.webSocketSubProtocolName, hv...)
			} else {
				ctx.Response.statusCode = http.StatusBadRequest
				return false
			}
		} else {
			hv, ok = ctx.Request.Header.Get(HeaderSecWebSocketProtocol)
			if ok && len(hv) > 0 {
				for _, v := range bytes.Split(hv, []byte(",")) {
					v = utils.InplaceTrimAsciiSpace(v)
					if _, ok := p.Subprotocols[utils.S(v)]; ok {
						subprotocol = v
						ctx.Request.webSocketSubProtocolName = append(ctx.Request.webSocketSubProtocolName, v...)
						break
					}
				}
			}
		}
	}

	var compress bool
	if p.EnableCompression {
		for _, hv := range ctx.Response.Header.GetAll(HeaderSecWebSocketExtensions) {
			if bytes.Contains(hv, utils.B(websocketExtCompress)) {
				compress = true
				break
			}
		}
	}

	res := &ctx.Response
	res.statusCode = http.StatusSwitchingProtocols
	res.Header.Append(HeaderConnection, upgradeStr)
	res.Header.Append(HeaderUpgrade, websocketStr)
	if len(subprotocol) > 0 {
		res.Header.Append(HeaderSecWebSocketProtocol, subprotocol)
	}
	res.Header.Append(HeaderSecWebSocketAccept, utils.B(websocket.ComputeAcceptKey(utils.S(key))))
	if compress {
		ctx.Request.webSocketShouldDoCompression = true
		res.Header.Append(HeaderSecWebSocketExtensions, utils.B(websocketExt))
	}
	_ = ctx.Send()
	return true
}

var websocketWriteBufferPool sync.Pool

func (p *WebSocketProtocol) Hijack(ctx *RequestCtx) *websocket.Conn {
	req := &ctx.Request
	ctx.hijacked = true
	return websocket.NewConn(
		ctx.conn, true, req.webSocketShouldDoCompression,
		p.ReadBufferSize, p.WriteBufferSize,
		&websocketWriteBufferPool, nil, nil,
	)
}

var wsp *WebSocketProtocol

func init() {
	wsp = &WebSocketProtocol{
		Subprotocols:      nil,
		EnableCompression: false,
		ReadBufferSize:    2048,
		WriteBufferSize:   2048,
	}
}

func SetWebSocketProtocol(p *WebSocketProtocol) {
	wsp = p
}

type WebSocketHandlerFunc func(ctx context.Context, req *Request, conn *websocket.Conn, subProtocolName string)

func wshToHandler(wsh WebSocketHandlerFunc) RequestHandler {
	return RequestHandlerFunc(func(ctx *RequestCtx) {
		p := ctx.UpgradeProtocol()
		if p != "websocket" {
			ctx.SetStatus(StatusBadRequest)
			return
		}
		if !wsp.Handshake(ctx) {
			return
		}
		wsh(ctx, &ctx.Request, wsp.Hijack(ctx), utils.S(ctx.Request.webSocketSubProtocolName))
	})
}
