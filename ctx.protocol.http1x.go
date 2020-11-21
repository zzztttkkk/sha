package suna

import (
	"bytes"
	"github.com/zzztttkkk/suna/internal"
	"time"
)

var httpVersion = []byte("http/")

func (ctx *RequestCtx) initRequest() {
	req := &ctx.Request

	ctx.Context = ctx.makeRequestCtx()
	ctx.reqTime = time.Now()
	ctx.bodySize = req.Header.ContentLength()
	ctx.bodyRemain = ctx.bodySize

	ind := bytes.IndexByte(req.rawPath, '?')
	if ind > 0 {
		req.Path = inplaceUnquote(req.rawPath[:ind])
		req.queryStatus = ind + 2
	} else {
		req.Path = inplaceUnquote(req.rawPath)
	}

	// todo may request body should be uncompress
	req.bodyBufferPtr = &ctx.buf
}

func (ctx *RequestCtx) onRequestHeaderLine() {
	key := internal.InplaceTrimAsciiSpace(ctx.cHKey)
	val := internal.InplaceTrimAsciiSpace(ctx.buf)
	ctx.Request.Header.Append(key, val)
}

func (ctx *RequestCtx) feedHttp1xReqData(data []byte, offset, end int) (int, HttpError) {
	var v byte

	switch ctx.status {
	case 0: // parse first line
		for offset < end {
			v = data[offset]
			offset++
			ctx.fLSize++
			if ctx.fLSize > ctx.protocol.MaxFirstLintSize {
				return 10001, ErrRequestUrlTooLong
			}
			if v > 126 || v < 10 {
				return 10002, ErrBadConnection
			}
			if v == '\n' {
				ctx.status++
				ctx.buf = ctx.buf[:0]
				if len(ctx.Request.rawPath) < 1 || ctx.Request.rawPath[0] != '/' { // empty path
					return 10003, ErrBadConnection
				}

				version := inplaceLowercase(ctx.Request.version)
				if !bytes.HasPrefix(version, httpVersion) { // http version
					return 10004, ErrBadConnection
				}
				ctx.Request.version = version[5:]
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
					ctx.Request.rawPath = append(ctx.Request.rawPath, v)
				case 2:
					ctx.Request.version = append(ctx.Request.version, v)
				default:
					return 10005, ErrBadConnection
				}
			}
			ctx.buf = append(ctx.buf, v)
		}
	case 1: // parse header line
		for offset < end {
			v = data[offset]
			offset++
			ctx.hSize++
			if ctx.hSize > ctx.protocol.MaxHeadersSize {
				return 10006, ErrRequestHeaderFieldsTooLarge
			}
			if v > 126 || v < 10 {
				return 10007, ErrBadConnection
			}

			if v == '\n' {
				if len(ctx.cHKey) < 1 { // allM header data read done
					ctx.status++
					return offset, nil
				}
				ctx.onRequestHeaderLine()
				ctx.cHKey = ctx.cHKey[:0]
				ctx.buf = ctx.buf[:0]
				return offset, nil
			}

			if v == '\r' {
				ctx.kvSep = false
				ctx.cHKeyDoUpper = true
				continue
			}

			if !ctx.kvSep {
				if v == ':' {
					ctx.kvSep = true
				} else {
					if ctx.cHKeyDoUpper {
						ctx.cHKeyDoUpper = false
						v = toUpperTable[v]
					}
					ctx.cHKey = append(ctx.cHKey, v)
					if v == '-' {
						ctx.cHKeyDoUpper = true
					}
				}
				continue
			}
			ctx.buf = append(ctx.buf, v)
		}
	case 2:
		if ctx.Context == nil {
			ctx.initRequest()
			if ctx.bodySize > ctx.protocol.MaxBodySize {
				return 10008, ErrRequestEntityTooLarge
			}
		}

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
