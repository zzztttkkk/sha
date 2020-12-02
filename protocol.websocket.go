package suna

import (
	"bytes"
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/websocket"
	"log"
	"net/http"
	"time"
)

type SubWebSocketProtocol interface {
	OnMessage(t int, data []byte)
}

type WebSocketProtocol struct {
	Subprotocols      map[string]SubWebSocketProtocol
	EnableCompression bool
}

var headerSecWebSocketVersion = []byte("Sec-WebSocket-Version")
var headerSecWebSocketExt = []byte("Sec-WebSocket-Extensions")
var headerSecWebsocketKey = []byte("Sec-WebSocket-Key")
var headerSecWebSocketSubp = []byte("Sec-Websocket-Protocol")
var websocketStr = []byte("websocket")

const (
	websocketExtCompress     = "permessage-deflate"
	headerSecWebSocketAccept = "Sec-WebSocket-Accept"
)

func (p *WebSocketProtocol) Handshake(ctx *RequestCtx) bool {
	version, _ := ctx.Request.Header.Get(headerSecWebSocketVersion)
	if len(version) != 2 || version[0] != '1' || version[1] != '3' {
		ctx.Response.statusCode = http.StatusBadRequest
		return false
	}
	if _, ok := ctx.Response.Header.Get(headerSecWebSocketExt); ok {
		log.Println("websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported")
		ctx.Response.statusCode = http.StatusInternalServerError
		return false
	}
	key, _ := ctx.Request.Header.Get(headerSecWebsocketKey)
	if len(key) < 1 {
		ctx.Response.statusCode = http.StatusBadRequest
		return false
	}

	var subprotocol []byte
	if len(p.Subprotocols) > 0 {
		hv, ok := ctx.Response.Header.Get(headerSecWebSocketSubp)
		if ok {
			if sp, ok := p.Subprotocols[internal.S(hv)]; ok {
				subprotocol = hv
				ctx.Request.wsSubP = sp
			} else {
				ctx.Response.statusCode = http.StatusBadRequest
				return false
			}
		} else {
			hv, ok = ctx.Request.Header.Get(headerSecWebSocketSubp)
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
		for _, hv := range ctx.Response.Header.GetAll(headerSecWebSocketExt) {
			if bytes.Contains(hv, internal.B(websocketExtCompress)) {
				compress = true
				break
			}
		}
	}

	res := &ctx.Response
	res.statusCode = http.StatusSwitchingProtocols
	res.Header.Append(headerConnection, upgradeStr)
	res.Header.Append(headerUpgrade, websocketStr)
	if len(subprotocol) > 0 {
		res.Header.Append(headerSecWebSocketSubp, subprotocol)
	}
	res.Header.AppendStr(headerSecWebSocketAccept, websocket.ComputeAcceptKey(internal.S(key)))
	if compress {
		ctx.Request.wsDoCompress = true
		res.Header.AppendStr(
			"Sec-WebSocket-Extensions",
			"permessage-deflate; server_no_context_takeover; client_no_context_takeover",
		)
	}
	_ = ctx.Send()
	return true
}

func (p *WebSocketProtocol) Hijack(ctx *RequestCtx) *websocket.Conn {
	req := &ctx.Request
	ctx.hijacked = true
	return websocket.NewConn(
		ctx.conn, true, req.wsDoCompress,
		0, 0, nil, nil, nil,
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
		if string(p) != "websocket" {
			_, _ = ctx.WriteString(fmt.Sprintf("Hello world: %s", time.Now()))
			return
		}
		if !wsp.Handshake(ctx) {
			return
		}
		wsh(ctx, &ctx.Request, wsp.Hijack(ctx))
	})
}
