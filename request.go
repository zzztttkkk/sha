package sha

import "github.com/zzztttkkk/sha/internal"

const _QueryParsed = -2

type Request struct {
	Header  Header
	Method  []byte
	_method _Method

	RawPath []byte
	Path    []byte
	qmIndex int // question mark index
	qmOK    bool
	Params  internal.Kvs

	cookies internal.Kvs
	query   Form
	body    Form
	files   FormFiles

	bodyStatus    int // 0: unparsed; 1: unsupported content type; 2: parsed
	cookieParsed  bool
	version       []byte
	bodyBufferPtr *[]byte

	// websocket
	wsSubP       SubWebSocketProtocol
	wsDoCompress bool
}

func (req *Request) Reset() {
	req.Header.Reset()
	req.Method = req.Method[:0]
	req.qmIndex = -1
	req.qmOK = false
	req.Path = req.Path[:0]
	req.Params.Reset()

	req.cookies.Reset()
	req.query.Reset()
	req.body.Reset()
	req.files = nil
	req.cookieParsed = false
	req.bodyStatus = 0
	req.RawPath = req.RawPath[:0]
	req.version = req.version[:0]
	req.bodyBufferPtr = nil
	req.wsSubP = nil
	req.wsDoCompress = false
}

func (req *Request) Cookie(key string) ([]byte, bool) {
	if !req.cookieParsed {
		v, ok := req.Header.Get(HeaderCookie)
		if ok {
			var key []byte
			var buf []byte

			for _, b := range v {
				switch b {
				case '=':
					key = append(key, buf...)
					buf = buf[:0]
				case ';':
					req.cookies.Set(internal.S(decodeURI(key)), decodeURI(buf))
					key = key[:0]
					buf = buf[:0]
				case ' ':
					continue
				default:
					buf = append(buf, b)
				}
			}
			req.cookies.Set(internal.S(decodeURI(key)), decodeURI(buf))
		}
		req.cookieParsed = true
	}
	return req.cookies.Get(key)
}

func (ctx *RequestCtx) Cookie(key string) ([]byte, bool) { return ctx.Request.Cookie(key) }
