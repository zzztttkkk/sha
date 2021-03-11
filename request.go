package sha

import (
	"github.com/zzztttkkk/sha/utils"
	"strconv"
)

const _QueryParsed = -2

type URLParams struct {
	utils.Kvs
}

func (up *URLParams) GetInt(name string, base int) (int64, bool) {
	v, ok := up.Get(name)
	if !ok {
		return 0, false
	}
	ret, err := strconv.ParseInt(utils.S(v), base, 64)
	if err != nil {
		return 0, false
	}
	return ret, true
}

type Request struct {
	Header  Header
	Method  []byte
	_method _Method

	RawPath           []byte
	Path              []byte
	questionMarkIndex int
	gotQuestionMark   bool
	URLParams         URLParams

	cookies utils.Kvs
	query   Form
	body    Form
	files   FormFiles

	bodyStatus    int // 0: unparsed; 1: unsupported content type; 2: parsed
	cookieParsed  bool
	version       []byte
	bodyBufferPtr *[]byte

	// websocket
	webSocketSubProtocolName     []byte
	webSocketShouldDoCompression bool
}

func (req *Request) Reset() {
	req.Header.Reset()
	req.Method = req.Method[:0]
	req.questionMarkIndex = 0
	req.gotQuestionMark = false
	req.Path = req.Path[:0]
	req.URLParams.Reset()

	req.cookies.Reset()
	req.query.Reset()
	req.body.Reset()
	req.files = nil
	req.cookieParsed = false
	req.bodyStatus = 0
	req.RawPath = req.RawPath[:0]
	req.Path = req.Path[:0]
	req.version = req.version[:0]
	req.bodyBufferPtr = nil
	req.webSocketSubProtocolName = req.webSocketSubProtocolName[:0]
	req.webSocketShouldDoCompression = false
}

func (req *Request) CookieValue(key string) ([]byte, bool) {
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
					req.cookies.Set(utils.S(decodeURI(key)), decodeURI(buf))
					key = key[:0]
					buf = buf[:0]
				case ' ':
					continue
				default:
					buf = append(buf, b)
				}
			}
			req.cookies.Set(utils.S(decodeURI(key)), decodeURI(buf))
		}
		req.cookieParsed = true
	}
	return req.cookies.Get(key)
}

func (req *Request) HeaderValue(key string) ([]byte, bool) { return req.Header.Get(key) }
