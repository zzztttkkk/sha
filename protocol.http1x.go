package sha

import (
	"bytes"
	"context"
	"github.com/zzztttkkk/sha/internal"
	"io"
	"net"
	"time"
)

type Http1xProtocol struct {
	Version                  []byte
	MaxRequestFirstLineSize  int
	MaxRequestHeaderPartSize int
	MaxRequestBodySize       int
	NewRequestContext        func(connCtx context.Context) context.Context

	IdleTimeout  time.Duration
	WriteTimeout time.Duration
	OnParseError func(conn net.Conn, err HttpError) (shouldCloseConn bool) // close connection if return true
	OnWriteError func(conn net.Conn, err error) (shouldCloseConn bool)     // close connection if return true

	ReadBufferSize     int
	MaxReadBufferSize  int
	MaxWriteBufferSize int

	server  *Server
	handler RequestHandler

	readBufferPool  *internal.FixedSizeBufferPool
	writeBufferPool *internal.BufferPool
}

var upgradeStr = []byte("upgrade")
var keepAliveStr = []byte("keep-alive")

const (
	closeStr  = "close"
	http11Str = "1.1"
)

func (protocol *Http1xProtocol) keepalive(ctx *RequestCtx) bool {
	req := &ctx.Request
	if string(req.version[5:]) < http11Str {
		return false
	}

	connVal, _ := ctx.Request.Header.Get(HeaderConnection)
	if string(inplaceLowercase(connVal)) == closeStr {
		return false
	}
	connVal, _ = ctx.Response.Header.Get(HeaderConnection)
	return string(inplaceLowercase(connVal)) != closeStr
}

var ZeroTime time.Time
var MaxIdleSleepTime = time.Millisecond * 100

func (protocol *Http1xProtocol) Serve(ctx context.Context, conn net.Conn) {
	var err error
	var n int
	var stop bool

	rctx := acquireRequestCtx()
	readBuf := protocol.readBufferPool.Get()
	rctx.Response.buf = protocol.writeBufferPool.Get()

	defer func() {
		protocol.readBufferPool.Put(readBuf)
		protocol.writeBufferPool.Put(rctx.Response.buf)
		ReleaseRequestCtx(rctx)
	}()

	rctx.conn = conn
	rctx.connTime = time.Now()
	rctx.protocol = protocol

	sleepDu := time.Millisecond * 10
	resetIdleTimeout := true

	for !stop {
		select {
		case <-ctx.Done():
			{
				return
			}
		default:
			if protocol.IdleTimeout > 0 && resetIdleTimeout {
				_ = conn.SetReadDeadline(time.Now().Add(protocol.IdleTimeout))
			}

			offset := 0
			n, err = conn.Read(readBuf.Data)
			if err != nil {
				if err == io.EOF {
					time.Sleep(sleepDu)
					sleepDu = sleepDu * 2
					resetIdleTimeout = false
					if sleepDu > MaxIdleSleepTime {
						sleepDu = MaxIdleSleepTime
					}
					continue
				}
				return
			}

			if protocol.IdleTimeout > 0 {
				_ = conn.SetReadDeadline(ZeroTime)
				resetIdleTimeout = true
			}

			for offset != n {
				offset, err = rctx.feedHttp1xReqData(readBuf.Data, offset, n)
				if err != nil {
					if protocol.OnParseError != nil {
						if protocol.OnParseError(conn, err.(HttpError)) {
							return
						}
					} else {
						return
					}
				}
			}

			if rctx.status == 2 && rctx.bodyRemain < 1 {
				if protocol.server.AutoCompress {
					rctx.AutoCompress()
				}

				rctx.Context = protocol.NewRequestContext(ctx)
				protocol.handler.Handle(rctx)

				if rctx.hijacked {
					return
				}

				if protocol.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(time.Now().Add(protocol.WriteTimeout))
				}

				if protocol.keepalive(rctx) {
					rctx.Response.Header.Set(HeaderConnection, keepAliveStr)
				} else {
					stop = true
				}

				if err := rctx.sendHttp1xResponseBuffer(); err != nil {
					if protocol.OnWriteError != nil {
						if protocol.OnWriteError(conn, err) {
							return
						}
					} else {
						return
					}
				}

				if protocol.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(ZeroTime)
				}

				rctx.Reset()
			}
		}
		continue
	}
}

var httpVersion = []byte("HTTP/")

func (ctx *RequestCtx) initRequest() {
	req := &ctx.Request

	ctx.reqTime = time.Now()

	ctx.bodyRemain = ctx.bodySize

	switch req.Method[0] {
	case 'G':
		if string(req.Method) == MethodGet {
			req._method = _MGet
		}
	case 'H':
		if string(req.Method) == MethodHead {
			req._method = _MHead
		}
	case 'P':
		switch string(req.Method) {
		case MethodPatch:
			req._method = _MPatch
		case MethodPost:
			req._method = _MPost
		case MethodPut:
			req._method = _MPut
		}
	case 'D':
		if string(req.Method) == MethodDelete {
			req._method = _MDelete
		}
	case 'C':
		if string(req.Method) == MethodConnect {
			req._method = _MConnect
		}
	case 'O':
		if string(req.Method) == MethodOptions {
			req._method = _MOptions
		}
	case 'T':
		if string(req.Method) == MethodTrace {
			req._method = _MTrace
		}
	}

	if req._method != _MConnect {
		if req.qmOK && req.qmIndex > 0 {
			req.Path = decodeURI(req.RawPath[:req.qmIndex])
		} else {
			req.Path = decodeURI(req.RawPath)
		}
		req.bodyBufferPtr = &ctx.buf
	}
}

func (ctx *RequestCtx) feedHttp1xReqData(data []byte, offset, end int) (int, HttpError) {
	var v byte
	req := &ctx.Request

	switch ctx.status {
	case 0: // parse first line
		for offset < end {
			v = data[offset]
			offset++
			ctx.firstLineSize++
			if ctx.firstLineSize > ctx.protocol.MaxRequestFirstLineSize {
				return 10001, ErrRequestUrlTooLong
			}

			if v == '\n' { // end of first line
				ctx.status++
				ctx.buf = ctx.buf[:0]
				if len(req.RawPath) < 1 { // empty path
					return 10003, ErrBadConnection
				}

				if len(req.version) < 8 && !bytes.HasPrefix(req.version, httpVersion) { // http version
					return 10004, ErrBadConnection
				}
				return offset, nil
			}

			switch v {
			case '\r':
				continue
			case ' ':
				ctx.fStatus += 1
			default:
				switch ctx.fStatus {
				case 0:
					req.Method = append(req.Method, toUpperTable[v])
				case 1:
					req.RawPath = append(req.RawPath, v)
					if !req.qmOK {
						if v == '?' {
							req.qmOK = true
						} else {
							req.qmIndex++
						}
					}
				case 2:
					req.version = append(req.version, toUpperTable[v])
				default:
					return 10005, ErrBadConnection
				}
			}
		}
	case 1: // parse header line
		for offset < end {
			v = data[offset]
			offset++
			ctx.headersSize++
			if ctx.headersSize > ctx.protocol.MaxRequestHeaderPartSize {
				return 10006, ErrRequestHeaderFieldsTooLarge
			}

			if v == '\n' {
				if len(ctx.currentHeaderKey) < 1 { // all header data read done
					ctx.status++
					ctx.bodySize = req.Header.ContentLength()
					if ctx.bodySize > ctx.protocol.MaxRequestBodySize {
						return 10008, ErrRequestEntityTooLarge
					}
					ctx.initRequest()
					return offset, nil
				}

				ctx.Request.Header.AppendBytes(
					internal.InplaceTrimAsciiSpace(ctx.currentHeaderKey),
					decodeURI(internal.InplaceTrimAsciiSpace(ctx.buf)),
				)
				ctx.currentHeaderKey = ctx.currentHeaderKey[:0]
				ctx.buf = ctx.buf[:0]
				return offset, nil
			}

			if v == '\r' {
				ctx.headerKVSepRead = false
				ctx.cHKeyDoUpper = true
				continue
			}

			if !ctx.headerKVSepRead {
				if v == ':' {
					ctx.headerKVSepRead = true
				} else {
					if ctx.cHKeyDoUpper {
						ctx.cHKeyDoUpper = false
						v = toUpperTable[v]
					}
					ctx.currentHeaderKey = append(ctx.currentHeaderKey, v)
					if v == '-' {
						ctx.cHKeyDoUpper = true
					}
				}
				continue
			}
			ctx.buf = append(ctx.buf, v)
		}
	case 2:
		size := end - offset
		if size > ctx.bodyRemain {
			return 10009, ErrBadConnection
		}
		ctx.buf = append(ctx.buf, data[offset:end]...)
		ctx.bodyRemain -= size
		return end, nil
	}
	return offset, nil
}
