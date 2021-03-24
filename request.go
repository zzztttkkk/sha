package sha

import (
	"github.com/zzztttkkk/sha/jsonx"
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
					req.cookies.Set(utils.S(utils.DecodeURI(key)), utils.DecodeURI(buf))
					key = key[:0]
					buf = buf[:0]
				case ' ':
					continue
				default:
					buf = append(buf, b)
				}
			}
			req.cookies.Set(utils.S(utils.DecodeURI(key)), utils.DecodeURI(buf))
		}
		req.cookieParsed = true
	}
	return req.cookies.Get(key)
}

func (req *Request) HeaderValue(key string) ([]byte, bool) { return req.Header.Get(key) }

func (req *Request) SetMethod(method string) *Request {
	req.Method = utils.B(method)
	return req
}

func (req *Request) SetPath(path []byte) *Request {
	req.Path = path
	return req
}

func (req *Request) SetPathString(path string) *Request {
	req.Path = utils.B(path)
	return req
}

func (req *Request) SetQuery(mvm MultiValueMap) *Request {
	for k, vl := range mvm {
		for _, v := range vl {
			req.query.AppendString(k, v)
		}
	}
	return req
}

func (req *Request) SetFormBody(mvm MultiValueMap) *Request {
	for k, vl := range mvm {
		for _, v := range vl {
			req.body.AppendString(k, v)
		}
	}
	req.Header.SetContentType(MIMEForm)
	return req
}

func (req *Request) SetJSONBody(v interface{}) *Request {
	req.Header.SetContentType(MIMEJson)
	b, e := jsonx.Marshal(v)
	if e != nil {
		panic(e)
	}
	req.bodyBufferPtr = &b
	return req
}

func (req *Request) SetRawBody(v []byte) *Request {
	req.bodyBufferPtr = &v
	return req
}
