package sha

import (
	"bytes"
	"context"
	"net/http"
	"sync"

	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/websocket"
)

type WebSocketOptions struct {
	ReadBufferSize    int
	WriteBufferSize   int
	EnableCompression bool
	SelectSubProtocol func(ctx *RequestCtx) string
}

var defaultWebSocketProtocolOption = WebSocketOptions{
	ReadBufferSize:    2048,
	WriteBufferSize:   2048,
	EnableCompression: true,
}

type _WebSocketProtocol struct {
	opt WebSocketOptions
}

func NewWebSocketProtocol(opt *WebSocketOptions) WebSocketProtocol {
	v := &_WebSocketProtocol{}
	if opt == nil {
		opt = &defaultWebSocketProtocolOption
	}
	v.opt = *opt
	return v
}

const (
	websocketStr         = "websocket"
	websocketExtCompress = "permessage-deflate"
	websocketExt         = "permessage-deflate; server_no_context_takeover; client_no_context_takeover"
)

func (p *_WebSocketProtocol) Handshake(ctx *RequestCtx) (string, bool, bool) {
	version, _ := ctx.Request.Header().Get(HeaderSecWebSocketVersion)
	if len(version) != 2 || version[0] != '1' || version[1] != '3' {
		ctx.Response.statusCode = http.StatusBadRequest
		return "", false, false
	}

	key, _ := ctx.Request.Header().Get(HeaderSecWebSocketKey)
	if len(key) < 1 {
		ctx.Response.statusCode = http.StatusBadRequest
		return "", false, false
	}

	var subProtocol string
	if p.opt.SelectSubProtocol != nil {
		subProtocol = p.opt.SelectSubProtocol(ctx)
	}

	var compress bool
	if p.opt.EnableCompression {
		for _, hv := range ctx.Response.Header().GetAll(HeaderSecWebSocketExtensions) {
			if bytes.Contains(hv, utils.B(websocketExtCompress)) {
				compress = true
				break
			}
		}
	}

	res := &ctx.Response
	res.statusCode = http.StatusSwitchingProtocols
	res.Header().AppendString(HeaderConnection, upgrade)
	res.Header().AppendString(HeaderUpgrade, websocketStr)
	res.Header().Append(HeaderSecWebSocketAccept, utils.B(websocket.ComputeAcceptKey(utils.S(key))))
	if compress {
		res.Header().Append(HeaderSecWebSocketExtensions, utils.B(websocketExt))
	}
	if len(subProtocol) > 0 {
		ctx.Response.header.SetString(HeaderSecWebSocketProtocol, subProtocol)
	}

	if err := sendResponse(ctx.w, &ctx.Response); err != nil {
		return "", false, false
	}
	return subProtocol, compress, true
}

var websocketWriteBufferPool sync.Pool

func (p *_WebSocketProtocol) Hijack(ctx *RequestCtx, subProtocol string, compress bool) *websocket.Conn {
	ctx.Hijack()
	return websocket.NewConnExt(
		ctx.conn, subProtocol, true, compress,
		p.opt.ReadBufferSize, p.opt.WriteBufferSize,
		&websocketWriteBufferPool, ctx.r, nil,
	)
}

type WebsocketHandlerFunc func(ctx context.Context, req *Request, conn *websocket.Conn)

func wshToHandler(wsh WebsocketHandlerFunc) RequestCtxHandler {
	return RequestCtxHandlerFunc(func(ctx *RequestCtx) {
		p := ctx.UpgradeProtocol()
		if p != websocketStr {
			ctx.Response.statusCode = StatusBadRequest
			return
		}
		serv := ctx.Value(CtxKeyServer).(*Server)
		wsp := serv.websocketProtocol
		if wsp == nil {
			ctx.Response.statusCode = StatusNotImplemented
			return
		}
		subProtocol, compress, ok := wsp.Handshake(ctx)
		if !ok {
			return
		}
		wsh(ctx, &ctx.Request, wsp.Hijack(ctx, subProtocol, compress))
	})
}
