package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
)

func (protocol *_Http11Protocol) sendResponseBuffer(ctx *RequestCtx) error {
	res := &ctx.Response
	if res.compressWriter != nil {
		err := res.compressWriter.Flush()
		if err != nil {
			return err
		}
	}

	size := int64(len(res.bodyBuf.Data))

	res.Header.SetContentLength(size)
	err := protocol.writeHeader(ctx)
	if err != nil {
		return err
	}

	_, err = res.sendBuf.Write(ctx.Response.bodyBuf.Data)
	if err != nil {
		return err
	}
	return res.sendBuf.Flush()
}

const (
	EndLine     = "\r\n"
	headerKVSep = ": "
)

var ErrUnknownResponseStatusCode = fmt.Errorf("sha: unknown response status code")

func (protocol *_Http11Protocol) writeHeader(ctx *RequestCtx) error {
	res := &ctx.Response

	if res.statusCode < 1 {
		res.statusCode = 200
	}

	statusTxt := statusTextMap[res.statusCode]
	if len(statusTxt) < 1 {
		return ErrUnknownResponseStatusCode
	}

	res.headerBuf = append(res.headerBuf, http11...)
	res.headerBuf = append(res.headerBuf, ' ')
	res.headerBuf = append(res.headerBuf, strconv.FormatInt(int64(res.statusCode), 10)...)
	res.headerBuf = append(res.headerBuf, ' ')
	res.headerBuf = append(res.headerBuf, statusTxt...)
	res.headerBuf = append(res.headerBuf, EndLine...)

	res.Header.EachItem(
		func(item *utils.KvItem) bool {
			res.headerBuf = append(res.headerBuf, item.Key...)
			res.headerBuf = append(res.headerBuf, headerKVSep...)
			encodeHeaderValue(item.Val, &res.headerBuf)
			res.headerBuf = append(res.headerBuf, EndLine...)
			return true
		},
	)

	res.headerBuf = append(res.headerBuf, EndLine...)
	_, e := res.sendBuf.Write(res.headerBuf)
	return e
}
