package sha

import (
	"bufio"
	"bytes"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"time"
)

func sendPocket(buf *bufio.Writer, pocket *_HTTPPocket) error {
	const (
		endLine     = "\r\n"
		headerKVSep = ": "
	)
	var contentLength int
	if pocket.body != nil {
		contentLength = pocket.body.Len()
	}
	pocket.header.SetContentLength(int64(contentLength))

	pocket.header.EachItem(
		func(item *utils.KvItem) bool {
			buf.Write(item.Key)
			buf.WriteString(headerKVSep)
			utils.EncodeHeaderValueToBuf(item.Val, buf)
			buf.WriteString(endLine)
			return true
		},
	)
	_, _ = buf.WriteString(endLine)

	if contentLength > 0 {
		_, _ = buf.Write(pocket.body.Bytes())
	}
	return buf.Flush()
}

func sendResponse(w *bufio.Writer, res *Response) error {
	if res.cw != nil {
		if err := res.cw.Flush(); err != nil {
			return err
		}
	}
	res.time = time.Now().UnixNano()

	const (
		endLine     = "\r\n"
		httpVersion = "HTTP/1.1"
		_200        = "200"
		ok          = "OK"
		unknown     = "Unknown Status Code"
	)

	if len(res.fl1) < 1 {
		w.WriteString(httpVersion)
	} else {
		w.Write(res.fl1)
	}
	w.WriteByte(' ')

	if len(res.fl2) < 1 {
		if res.statusCode == 0 {
			w.WriteString(_200)
		} else {
			w.WriteString(strconv.FormatInt(int64(res.statusCode), 10))
		}
	} else {
		w.Write(res.fl2)
	}
	w.WriteByte(' ')

	if len(res.fl3) < 1 {
		if res.statusCode == 0 {
			w.WriteString(ok)
		} else {
			v := statusTextMap[res.statusCode]
			if len(v) < 1 {
				w.WriteString(unknown)
			} else {
				w.Write(v)
			}
		}
	} else {
		w.Write(res.fl3)
	}
	w.WriteString(endLine)

	return sendPocket(w, &res._HTTPPocket)
}

func sendRequest(w *bufio.Writer, req *Request) error {
	req.time =  time.Now().UnixNano()

	const (
		endLine     = "\r\n"
		httpVersion = "HTTP/1.1"
	)

	if len(req.fl1) < 1 {
		w.WriteString(MethodGet)
	} else {
		w.Write(req.fl1)
	}
	w.WriteByte(' ')

	q := false
	if len(req.fl2) < 1 {
		w.WriteByte('/')
	} else {
		w.Write(req.fl2)
		q = bytes.IndexByte(req.fl2, '?') > -1
	}

	qs := req.query.Size()
	if qs > 0 {
		if q {
			w.WriteByte('&')
		} else {
			w.WriteByte('?')
		}
		req.query.EncodeToBuf(w)
	}

	w.WriteByte(' ')

	if len(req.fl3) < 1 {
		w.WriteString(httpVersion)
	} else {
		w.Write(req.fl3)
	}
	w.WriteString(endLine)

	if req.bodyForm.Size() > 0 {
		_, _ = req._HTTPPocket.Write(nil)
		req.bodyForm.EncodeToBuf(req.body)
	}
	return sendPocket(w, &req._HTTPPocket)
}
