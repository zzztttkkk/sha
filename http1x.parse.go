package sha

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"time"
	"unicode"
)

var numMap []bool

func init() {
	numMap = make([]bool, 256)
	for _, b := range "" {
		numMap[b] = true
	}
}

var ErrBadHTTPPocketData = errors.New("sha.http: bad http pocket data")
var ErrTimeout = errors.New("sha.http: timeout")

func parsePocket(ctx context.Context, reader *bufio.Reader, readBuf []byte, pocket *_HTTPPocket, opt *HTTPOption) error {
	var (
		skipNewLine bool
		skipSpace   bool
		headerItem  *utils.KvItem
		keySep      bool
		keyDone     bool
		bodyRemain  int
		parseStatus int

		firstLineSize int
		headerSize    int
		b             byte
		e             error
	)
	pocket.header.fromOutSide = true

	for {
		if parseStatus == 4 {
			if bodyRemain == 0 {
				contentLength := pocket.header.ContentLength()
				if contentLength < 1 {
					return nil
				}
				bodyRemain = contentLength
				if opt.MaxBodySize > 0 && contentLength > opt.MaxBodySize {
					return ErrBadHTTPPocketData
				}
			}

			l, e := reader.Read(readBuf)
			if e != nil {
				return e
			}
			if l == 0 {
				goto checkCtx
			}

			_, _ = pocket.Write(readBuf[:l])

			bodyRemain -= l
			if bodyRemain == 0 {
				return nil
			}
			if bodyRemain < 0 {
				return ErrBadHTTPPocketData
			}

			goto checkCtx
		}

		b, e = reader.ReadByte()
		if e != nil {
			return e
		}

		if skipNewLine {
			if b != '\n' {
				return ErrBadHTTPPocketData
			}
			skipNewLine = false
			continue
		}
		if skipSpace {
			if b != ' ' {
				return ErrBadHTTPPocketData
			}
			skipSpace = false
			continue
		}

		switch parseStatus {
		case 0: // req method or res version
			if b == ' ' {
				parseStatus++
				continue
			}
			pocket.fl1 = append(pocket.fl1, toUpperTable[b])
		case 1: // req path or res status code
			if b == ' ' {
				parseStatus++
				continue
			}
			pocket.fl2 = append(pocket.fl2, b)
		case 2: // req version or res status phrase
			if b == '\r' {
				skipNewLine = true
				parseStatus++

				goto checkCtx
			}
			pocket.fl3 = append(pocket.fl3, toUpperTable[b])
		case 3:
			if !keySep && b == ':' {
				keySep = true
				skipSpace = true
				keyDone = true
				continue
			}

			if b == '\r' {
				keySep = false
				keyDone = false
				skipNewLine = true

				if headerItem == nil {
					parseStatus++
					b, e := reader.ReadByte()
					if e != nil {
						return e
					}
					if b != '\n' {
						return ErrBadHTTPPocketData
					}
					continue
				}

				// change key to lower in the safe way
				header := &pocket.header
				header.buf.Reset()
				key := utils.S(headerItem.Key)
				for _, ru := range key {
					if ru > 255 {
						header.utf8Key = true
					}
					header.buf.WriteRune(unicode.ToLower(ru))
				}
				headerItem.Key = headerItem.Key[:0]
				headerItem.Key = append(headerItem.Key, header.buf.String()...)

				headerItem = nil
				goto checkCtx
			}

			if headerItem == nil {
				headerItem = pocket.header.AppendBytes(nil, nil)
			}

			if keyDone {
				headerItem.Val = append(headerItem.Val, b)
			} else {
				headerItem.Key = append(headerItem.Key, b)
			}
		}

		if opt.MaxFirstLineSize > 0 && parseStatus < 3 {
			firstLineSize++
			if firstLineSize > opt.MaxFirstLineSize {
				return ErrBadHTTPPocketData
			}
		}

		if opt.MaxHeaderPartSize > 0 && parseStatus == 3 {
			headerSize++
			if headerSize > opt.MaxHeaderPartSize {
				return ErrBadHTTPPocketData
			}
		}

	checkCtx:
		select {
		case <-ctx.Done():
			return ErrTimeout
		default:
		}
	}

}

var httpVersionPrefix = []byte("HTTP/")

func parseRequest(ctx context.Context, r *bufio.Reader, buf []byte, req *Request, opt *HTTPOption) error {
	if err := parsePocket(ctx, r, buf, &req._HTTPPocket, opt); err != nil {
		return err
	}
	if len(req.Method()) < 1 {
		return ErrBadHTTPPocketData
	}
	if len(req.HTTPVersion()) != 8 || !bytes.HasPrefix(req.HTTPVersion(), httpVersionPrefix) {
		return ErrBadHTTPPocketData
	}
	req.methodToEnum()
	req.parsePath()
	req.time = time.Now().UnixNano()
	return nil
}

func parseResponse(ctx context.Context, r *bufio.Reader, buf []byte, res *Response, opt *HTTPOption) error {
	if err := parsePocket(ctx, r, buf, &res._HTTPPocket, opt); err != nil {
		return err
	}
	sv, err := strconv.ParseInt(utils.S(res.fl2), 10, 64)
	if err != nil {
		return ErrBadHTTPPocketData
	}
	res.statusCode = int(sv)

	if len(res.HTTPVersion()) != 8 || !bytes.HasPrefix(res.HTTPVersion(), httpVersionPrefix) {
		return ErrBadHTTPPocketData
	}
	if len(res.Phrase()) < 1 {
		return ErrBadHTTPPocketData
	}
	res.time = time.Now().UnixNano()
	return nil
}
