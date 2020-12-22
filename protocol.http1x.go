package sha

import (
	"bufio"
	"bytes"
	"context"
	"github.com/zzztttkkk/sha/internal"
	"net"
	"sync"
	"time"
)

type Http1xProtocol struct {
	Version                  []byte
	MaxRequestFirstLineSize  int
	MaxRequestHeaderPartSize int // all header lines
	MaxRequestBodySize       int

	OnParseError func(conn net.Conn, err HttpError) (shouldCloseConn bool) // close connection if return true
	OnWriteError func(conn net.Conn, err error) (shouldCloseConn bool)     // close connection if return true

	ReadBufferSize                int
	MaxReadBufferSize             int
	MaxResponseBodyBufferSize     int
	DefaultResponseSendBufferSize int

	server  *Server
	handler RequestHandler

	readBufferPool    *internal.FixedSizeBufferPool
	resBodyBufferPool *internal.BufferPool
	resSendBufferPool *sync.Pool
}

var upgradeStr = []byte("upgrade")
var keepAliveStr = []byte("keep-alive")

const (
	closeStr  = "close"
	http11Str = "1.1"
	keepAlive = "keep-alive"
)

func (protocol *Http1xProtocol) keepalive(ctx *RequestCtx) bool {
	connVal, _ := ctx.Response.Header.Get(HeaderConnection) // disable keep-alive by response
	if string(inplaceLowercase(connVal)) == closeStr {
		return false
	}
	connVal, _ = ctx.Request.Header.Get(HeaderConnection) // disable keep-alive by request
	connValS := internal.S(connVal)
	if connValS == closeStr {
		return false
	}
	if connValS == keepAlive { // enable keep-alive by request
		return true
	}
	return string(ctx.Request.version[5:]) >= http11Str // if http version >= 1.1, enable keep-alive default
}

var zeroTime time.Time

func (protocol *Http1xProtocol) Serve(ctx context.Context, conn net.Conn) {
	var err error
	var n int
	var keepAlive = true
	var cancelFn func()

	rctx := acquireRequestCtx()
	rctx.isTLS = protocol.server.isTls
	readBuf := protocol.readBufferPool.Get()
	rctx.Response.bodyBuf = protocol.resBodyBufferPool.Get()
	bufI := protocol.resSendBufferPool.Get()
	if bufI == nil {
		rctx.Response.sendBuf = bufio.NewWriterSize(conn, protocol.DefaultResponseSendBufferSize)
	} else {
		rctx.Response.sendBuf = bufI.(*bufio.Writer)
		rctx.Response.sendBuf.Reset(conn)
	}

	defer func() {
		protocol.readBufferPool.Put(readBuf)
		protocol.resBodyBufferPool.Put(rctx.Response.bodyBuf)
		protocol.resSendBufferPool.Put(rctx.Response.sendBuf)
		ReleaseRequestCtx(rctx)
	}()

	rctx.conn = conn
	rctx.connTime = time.Now()
	rctx.protocol = protocol

	idleTimeout := protocol.server.IdleTimeout
	readTimeout := protocol.server.ReadTimeout
	writeTimeout := protocol.server.WriteTimeout
	autoCompression := protocol.server.AutoCompression

	inIdle := false

	if readTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	}

	for keepAlive {
		offset := 0
		n, err = conn.Read(readBuf.Data)
		if n < 1 {
			if err == nil {
				continue
			}
			// `net.Conn.Read` is a blocking call, got 'io.EOF' means that the client closes this connection.
			return
		}

		if inIdle { // got data, stop idle, reset ReadTimeout
			inIdle = false
			_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		}

		// consume all the buffered data
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

		if rctx.status != 2 || rctx.bodyRemain > 0 {
			continue
		}

		// got a http1x request
		if autoCompression {
			rctx.AutoCompress()
		}

		rctx.Context, cancelFn = context.WithCancel(ctx)
		protocol.handler.Handle(rctx)

		if rctx.hijacked {
			cancelFn()
			return
		}

		if writeTimeout > 0 {
			_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		}

		if protocol.keepalive(rctx) {
			rctx.Response.Header.Set(HeaderConnection, keepAliveStr)
		} else {
			keepAlive = false
		}

		if err := rctx.sendHttp1xResponseBuffer(); err != nil {
			cancelFn()
			if protocol.OnWriteError != nil {
				if protocol.OnWriteError(conn, err) {
					return
				}
			} else {
				return
			}
		}

		if writeTimeout > 0 {
			_ = conn.SetWriteDeadline(zeroTime)
		}

		inIdle = true
		if idleTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(idleTimeout))
		} else if readTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		} else {
			_ = conn.SetReadDeadline(zeroTime)
		}

		cancelFn()
		rctx.Reset()
	}
}

var httpVersion = []byte("HTTP/")

func cleanPath(p []byte) []byte {
	ignoreSlash := false
	ind := 0
	for _, b := range p {
		if b == '/' {
			if !ignoreSlash {
				p[ind] = b
				ind++
				ignoreSlash = true
			}
		} else {
			ignoreSlash = false
			p[ind] = b
			ind++
		}
	}
	return p[:ind]
}

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
		if req.gotQuestionMark {
			req.questionMarkIndex--
			req.Path = append(req.Path, req.RawPath[:req.questionMarkIndex]...)
		} else {
			req.Path = append(req.Path, req.RawPath...)
		}
		if len(req.Path) < 1 {
			req.Path = append(req.Path, '/')
		}

		req.Path = decodeURI(cleanPath(req.Path))
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

			// ascii
			if v > 127 {
				return 10002, ErrBadConnection
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
				ctx.firstLineStatus += 1
			default:
				switch ctx.firstLineStatus {
				case 0:
					req.Method = append(req.Method, toUpperTable[v])
				case 1:
					req.RawPath = append(req.RawPath, v)
					if !req.gotQuestionMark {
						req.questionMarkIndex++
						if v == '?' {
							req.gotQuestionMark = true
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

			// ascii
			//if v > 127 {
			//	return 10007, ErrBadConnection
			//}

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
