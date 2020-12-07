package suna

import (
	"bytes"
	"context"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/websocket"
	"log"
	"net/http"
)

type SubWebSocketProtocol interface {
	OnMessage(t int, data []byte)
}

type WebSocketProtocol struct {
	Subprotocols      map[string]SubWebSocketProtocol
	EnableCompression bool
}

var websocketStr = []byte("websocket")

const (
	websocketExtCompress = "permessage-deflate"
	websocketExt         = "permessage-deflate; server_no_context_takeover; client_no_context_takeover"
)

func (p *WebSocketProtocol) Handshake(ctx *RequestCtx) bool {
	version, _ := ctx.Request.Header.Get(internal.B(HeaderSecWebSocketVersion))
	if len(version) != 2 || version[0] != '1' || version[1] != '3' {
		ctx.Response.statusCode = http.StatusBadRequest
		return false
	}
	if _, ok := ctx.Response.Header.Get(internal.B(HeaderSecWebSocketExtensions)); ok {
		log.Println("websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported")
		ctx.Response.statusCode = http.StatusInternalServerError
		return false
	}
	key, _ := ctx.Request.Header.Get(internal.B(HeaderSecWebSocketKey))
	if len(key) < 1 {
		ctx.Response.statusCode = http.StatusBadRequest
		return false
	}

	var subprotocol []byte
	if len(p.Subprotocols) > 0 {
		hv, ok := ctx.Response.Header.Get(internal.B(HeaderSecWebSocketProtocol))
		if ok {
			if sp, ok := p.Subprotocols[internal.S(hv)]; ok {
				subprotocol = hv
				ctx.Request.wsSubP = sp
			} else {
				ctx.Response.statusCode = http.StatusBadRequest
				return false
			}
		} else {
			hv, ok = ctx.Request.Header.Get(internal.B(HeaderSecWebSocketProtocol))
			if ok && len(hv) > 0 {
				for _, v := range bytes.Split(hv, []byte(",")) {
					v = internal.InplaceTrimAsciiSpace(v)
					if sp, ok := p.Subprotocols[internal.S(v)]; ok {
						subprotocol = v
						ctx.Request.wsSubP = sp
						break
					}
				}
			}
		}
	}

	var compress bool
	if p.EnableCompression {
		for _, hv := range ctx.Response.Header.GetAll(internal.B(HeaderSecWebSocketExtensions)) {
			if bytes.Contains(hv, internal.B(websocketExtCompress)) {
				compress = true
				break
			}
		}
	}

	res := &ctx.Response
	res.statusCode = http.StatusSwitchingProtocols
	res.Header.Append(internal.B(HeaderConnection), upgradeStr)
	res.Header.Append(internal.B(HeaderUpgrade), websocketStr)
	if len(subprotocol) > 0 {
		res.Header.Append(internal.B(HeaderSecWebSocketProtocol), subprotocol)
	}
	res.Header.AppendStr(HeaderSecWebSocketAccept, websocket.ComputeAcceptKey(internal.S(key)))
	if compress {
		ctx.Request.wsDoCompress = true
		res.Header.AppendStr(HeaderSecWebSocketExtensions, websocketExt)
	}
	_ = ctx.Send()
	return true
}

func (p *WebSocketProtocol) Hijack(ctx *RequestCtx) *websocket.Conn {
	req := &ctx.Request
	ctx.hijacked = true
	return websocket.NewConn(
		ctx.conn, true, req.wsDoCompress, 0, 0, nil, nil, nil,
	)
}

var wsp *WebSocketProtocol

func init() {
	wsp = &WebSocketProtocol{
		Subprotocols:      nil,
		EnableCompression: false,
	}
}

func SetWebSocketProtocol(p *WebSocketProtocol) {
	wsp = p
}

type WebSocketHandlerFunc func(ctx context.Context, req *Request, conn *websocket.Conn)

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
		wsh(ctx, &ctx.Request, wsp.Hijack(ctx))
	})
}
