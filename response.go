package sha

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"sync"
)

type Response struct {
	_HTTPPocket
	statusCode int
	cw         _CompressionWriter
	cwp        *sync.Pool
}

func (res *Response) StatusCode() int { return res.statusCode }

func (res *Response) Phrase() string { return utils.S(res.fl3) }

func (res *Response) HTTPVersion() []byte { return res.fl1 }

func (res *Response) SetHTTPVersion(v string) *Response {
	res.fl1 = res.fl1[:0]
	res.fl1 = append(res.fl1, v...)
	return res
}

var ErrUnknownResponseStatusCode = fmt.Errorf("sha: unknown response status code")

func (res *Response) SetStatusCode(v int) *Response {
	res.statusCode = v
	res.fl3 = statusTextMap[v]
	if len(res.fl3) < 1 {
		panic(ErrUnknownResponseStatusCode)
	}
	return res
}

func (res *Response) Body() *bytes.Buffer { return res.body }

func (res *Response) Write(p []byte) (int, error) {
	if res.cw != nil {
		return res.cw.Write(p)
	}
	return res._HTTPPocket.Write(p)
}

func (res *Response) reset() {
	res._HTTPPocket.reset()
	res.statusCode = 0
	if res.cw != nil {
		res.cw.Reset(nil)
		res.cwp.Put(res.cw)
		res.cw = nil
		res.cwp = nil
	}
}

func (res *Response) ResetBody() {
	if res.body != nil {
		res.body.Reset()
	}
	if res.cw != nil {
		res.cw.Reset(&res._HTTPPocket)
	}
}
