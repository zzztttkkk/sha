package sha

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
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
var ErrCanceled = errors.New("sha.http: canceled")

func init() {
	internal.ErrorStatusByValue[ErrBadHTTPPocketData] = StatusBadRequest
	internal.ErrorStatusByValue[ErrCanceled] = StatusBadRequest
}

/* parsePocket read http pocket from `reader`, use fixed-size buffer `readBuf`
parseStatus:
0  --  first line first part
1  --  first line second part
2  --  first line third part
3  --  header lines
4  --  read fixed size body, `content-length`
5  --  read chunked body
*/
func parsePocket(ctx context.Context, reader *bufio.Reader, readBuf []byte, pocket *_HTTPPocket, opt *HTTPOptions) error {
	const (
		chunked = "chunked"
	)

	var (
		skipNewLine bool
		skipSpace   bool
		headerItem  *utils.KvItem
		keySep      bool
		keyDone     bool
		bodyRemain  int = -1
		parseStatus int

		firstLineSize int
		headerSize    int
		b             byte
		e             error
	)
	pocket.header.fromOutSide = true

	for {
		if parseStatus < 4 {
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
		}

		switch parseStatus {
		// req method or res version
		case 0:
			firstLineSize++
			if b == ' ' {
				parseStatus++
				goto checkCtxAndFirstLineSize
			}
			pocket.fl1 = append(pocket.fl1, toUpperTable[b])
			goto checkCtxAndFirstLineSize
		// req path or res status code
		case 1:
			firstLineSize++
			if b == ' ' {
				parseStatus++
				goto checkCtxAndFirstLineSize
			}
			pocket.fl2 = append(pocket.fl2, b)
			goto checkCtxAndFirstLineSize
		// req version or res status phrase
		case 2:
			firstLineSize++
			if b == '\r' {
				skipNewLine = true
				parseStatus++

				goto checkCtxAndFirstLineSize
			}
			pocket.fl3 = append(pocket.fl3, toUpperTable[b])
			goto checkCtxAndFirstLineSize
		// headers
		case 3:
			headerSize++
			if opt.MaxHeaderPartSize > 0 && headerSize > opt.MaxHeaderPartSize {
				return StatusError(StatusRequestHeaderFieldsTooLarge)
			}

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

				if headerItem == nil { // header done, get content-length
					parseStatus++

					b, e := reader.ReadByte()
					if e != nil {
						return e
					}
					if b != '\n' {
						return ErrBadHTTPPocketData
					}

					rn, _ := pocket.header.Get(HeaderTransferEncoding)
					// multi-values such as `chunked, gzip` is not supported. i think the `gzip` should be set to `Content-Encoding`
					if string(rn) == chunked {
						parseStatus++
						goto checkCtx
					}
					contentLength := pocket.header.ContentLength()
					if contentLength < 1 {
						return nil
					}
					bodyRemain = contentLength
					if opt.MaxBodySize > 0 && contentLength > opt.MaxBodySize {
						return StatusError(StatusRequestEntityTooLarge)
					}
					goto checkCtx
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
			continue
		// fixed size body
		case 4:
			for {
				if len(readBuf) > bodyRemain {
					readBuf = readBuf[:bodyRemain]
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
				goto checkCtx
			}
		// chunked body
		case 5:
			for {
				if bodyRemain < 0 {
					var line []byte
					line, _, e = reader.ReadLine()
					if e != nil {
						return e
					}
					if len(line) == 0 {
						goto checkCtx
					}
					var l int64
					l, e = strconv.ParseInt(utils.S(line), 16, 32)
					if e != nil {
						return e
					}
					bodyRemain = int(l)
					if bodyRemain == 0 {
						line, _, e = reader.ReadLine()
						if e != nil {
							return e
						}
						return nil
					}
					goto checkCtx
				}

				if len(readBuf) > bodyRemain {
					readBuf = readBuf[:bodyRemain]
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
					bodyRemain = -1
					goto checkCtx
				}
				goto checkCtx
			}
		}
		continue

	checkCtxAndFirstLineSize:
		if opt.MaxFirstLineSize > 0 && firstLineSize > opt.MaxFirstLineSize {
			return StatusError(StatusRequestURITooLong)
		}

	checkCtx:
		select {
		case <-ctx.Done():
			return ErrCanceled
		default:
		}
	}
}

var httpVersionPrefix = []byte("HTTP/")

func parseRequest(ctx context.Context, r *bufio.Reader, buf []byte, req *Request, opt *HTTPOptions) error {
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
	req.setTime()
	return nil
}

func parseResponse(ctx context.Context, r *bufio.Reader, buf []byte, res *Response, opt *HTTPOptions) error {
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
	res.setTime()
	return nil
}
