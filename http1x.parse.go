package sha

import (
	"bufio"
	"context"
	"errors"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"time"
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

func parsePocket(ctx context.Context, reader *bufio.Reader, readBuf []byte, pocket *_HTTPPocket) error {
	var (
		skipNewLine bool
		skipSpace   bool
		headerItem  *utils.KvItem
		keySep      bool
		keyDone     bool
		bodyRemain  int
		deadline    time.Time
		checkTime   bool
	)
	deadline, checkTime = ctx.Deadline()

	pocket.header.fromOutSide = true

	for {
		if checkTime && time.Now().After(deadline) {
			return ErrTimeout
		}

		if pocket.parseStatus == 4 {
			if bodyRemain == 0 {
				contentLength := pocket.header.ContentLength()
				if contentLength < 1 {
					return nil
				}
				bodyRemain = contentLength
			}

			l, e := reader.Read(readBuf)
			if e != nil {
				return e
			}
			if l == 0 {
				continue
			}

			_, _ = pocket.Write(readBuf[:l])

			bodyRemain -= l
			if bodyRemain == 0 {
				return nil
			}
			if bodyRemain < 0 {
				return ErrBadHTTPPocketData
			}
			continue
		}

		b, e := reader.ReadByte()
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

		switch pocket.parseStatus {
		case 0:
			if b == ' ' {
				pocket.parseStatus++
				continue
			}
			pocket.fl1 = append(pocket.fl1, toUpperTable[b])
		case 1:
			if b == ' ' {
				pocket.parseStatus++
				continue
			}
			pocket.fl2 = append(pocket.fl2, b)
		case 2:
			if b == '\r' {
				skipNewLine = true
				pocket.parseStatus++
				continue
			}
			pocket.fl3 = append(pocket.fl3, b)
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
					pocket.parseStatus++
					b, e := reader.ReadByte()
					if e != nil {
						return e
					}
					if b != '\n' {
						return ErrBadHTTPPocketData
					}
				}
				headerItem = nil
				continue
			}

			if headerItem == nil {
				headerItem = pocket.header.AppendBytes(nil, nil)
			}

			if keyDone {
				headerItem.Val = append(headerItem.Val, b)
			} else {
				headerItem.Key = append(headerItem.Key, toLowerTable[b])
			}
		}
	}
}

func parseRequest(ctx context.Context, r *bufio.Reader, buf []byte, req *Request) error {
	if err := parsePocket(ctx, r, buf, &req._HTTPPocket); err != nil {
		return err
	}
	req.methodToEnum()
	req.parsePath()
	return nil
}

func parseResponse(ctx context.Context, r *bufio.Reader, buf []byte, res *Response) error {
	if err := parsePocket(ctx, r, buf, &res._HTTPPocket); err != nil {
		return err
	}
	sv, err := strconv.ParseInt(utils.S(res.fl2), 10, 64)
	if err != nil {
		return ErrBadHTTPPocketData
	}
	res.statusCode = int(sv)
	return nil
}
