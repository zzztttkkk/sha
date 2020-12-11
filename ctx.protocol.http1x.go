package sha

import (
	"bytes"
	"github.com/zzztttkkk/sha/internal"
	"time"
)

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

	if ctx.Request._method != _MConnect {
		ind := bytes.IndexByte(req.RawPath, '?')
		if ind > 0 {
			req.Path = decodeURI(req.RawPath[:ind])
			req.queryStatus = ind + 2
		} else {
			req.Path = decodeURI(req.RawPath)
		}
		req.bodyBufferPtr = &ctx.buf
	}
}

func (ctx *RequestCtx) onRequestHeaderLine() {
	key := decodeURI(internal.InplaceTrimAsciiSpace(ctx.currentHeaderKey))
	val := decodeURI(internal.InplaceTrimAsciiSpace(ctx.buf))
	ctx.Request.Header.Append(key, val)
}

func (ctx *RequestCtx) feedHttp1xReqData(data []byte, offset, end int) (int, HttpError) {
	var v byte

	switch ctx.status {
	case 0: // parse first line
		for offset < end {
			v = data[offset]
			offset++
			ctx.firstLineSize++
			if ctx.firstLineSize > ctx.protocol.MaxRequestFirstLineSize {
				return 10001, ErrRequestUrlTooLong
			}
			if v > 126 || v < 10 { // just ascii
				return 10002, ErrBadConnection
			}
			if v == '\n' { // end of first line
				ctx.status++
				ctx.buf = ctx.buf[:0]
				if len(ctx.Request.RawPath) < 1 { // empty path
					return 10003, ErrBadConnection
				}

				if len(ctx.Request.version) < 8 && !bytes.HasPrefix(ctx.Request.version, httpVersion) { // http version
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
					ctx.Request.Method = append(ctx.Request.Method, toUpperTable[v])
				case 1:
					ctx.Request.RawPath = append(ctx.Request.RawPath, v)
				case 2:
					ctx.Request.version = append(ctx.Request.version, toUpperTable[v])
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
					ctx.bodySize = ctx.Request.Header.ContentLength()
					if ctx.bodySize > ctx.protocol.MaxRequestBodySize {
						return 10008, ErrRequestEntityTooLarge
					}
					ctx.initRequest()
					return offset, nil
				}
				ctx.onRequestHeaderLine()
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
					if v < 127 && ctx.cHKeyDoUpper {
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
