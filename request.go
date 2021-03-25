package sha

import (
	"bytes"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"time"
)

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

type URL struct {
	ok    bool
	Host  []byte
	Port  []byte
	Path  []byte
	Query []byte
}

type Request struct {
	_HTTPPocket
	reqTime time.Time

	_method _Method

	URLParams   URLParams
	url         URL
	queryParsed bool

	cookies  utils.Kvs
	query    Form
	bodyForm Form
	files    FormFiles

	bodyStatus   int // 0: unparsed; 1: unsupported content type; 2: parsed
	cookieParsed bool
	version      []byte

	// websocket
	webSocketSubProtocolName     []byte
	webSocketShouldDoCompression bool
}

func (req *Request) Reset() {
	req._HTTPPocket.reset()
	req._method = 0

	req.URLParams.Reset()
	req.url.ok = false

	req.reqTime = time.Time{}
	req.cookies.Reset()
	req.query.Reset()
	req.bodyForm.Reset()
	req.files = nil
	req.cookieParsed = false
	req.bodyStatus = 0
	req.version = req.version[:0]
	req.webSocketSubProtocolName = req.webSocketSubProtocolName[:0]
	req.webSocketShouldDoCompression = false
}

func (req *Request) Method() []byte { return req.fl1 }

func (req *Request) RawPath() []byte { return req.fl2 }

const defaultPath = "/"

func (req *Request) Path() []byte {
	req.parsePath()
	if len(req.url.Path) < 1 {
		return utils.B(defaultPath)
	}
	return req.url.Path
}

func (req *Request) Version() []byte { return req.fl3 }

func (req *Request) Body() *bytes.Buffer { return req._HTTPPocket.body }

func (req *Request) SetMethod(method string) *Request {
	req.fl1 = req.fl1[:0]
	req.fl1 = append(req.fl1, method...)
	return req
}

func (req *Request) SetVersion(version string) *Request {
	req.fl3 = req.fl3[:0]
	req.fl3 = append(req.fl3, version...)
	return req
}

func (req *Request) SetPath(path []byte) *Request {
	req.fl2 = req.fl2[:0]
	req.fl2 = append(req.fl2, path...)
	return req
}

func (req *Request) SetPathString(path string) *Request {
	req.fl2 = req.fl2[:0]
	req.fl2 = append(req.fl2, path...)
	return req
}

func (req *Request) CookieValue(key string) ([]byte, bool) {
	if !req.cookieParsed {
		v, ok := req.Header().Get(HeaderCookie)
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

func (req *Request) SetJSONBody(v interface{}) error {
	req.Header().SetContentType(MIMEJson)
	return jsonx.NewEncoder(&req._HTTPPocket).Encode(v)
}

func (req *Request) SetMultiValueMapBody(fv MultiValueMap) *Request {
	req.header.SetContentType(MIMEForm)
	req.bodyForm.LoadMap(fv)
	return req
}

func (req *Request) SetFormBody(form *Form) *Request {
	req.header.SetContentType(MIMEForm)
	req.bodyForm.LoadForm(form)
	return req
}

func (req *Request) methodToEnum() {
	if req._method != _MUnknown {
		return
	}

	method := utils.S(req.fl1)
	switch req.fl1[0] {
	case 'G':
		if method == MethodGet {
			req._method = _MGet
		}
	case 'H':
		if method == MethodHead {
			req._method = _MHead
		}
	case 'P':
		switch method {
		case MethodPatch:
			req._method = _MPatch
		case MethodPost:
			req._method = _MPost
		case MethodPut:
			req._method = _MPut
		}
	case 'D':
		if method == MethodDelete {
			req._method = _MDelete
		}
	case 'C':
		if method == MethodConnect {
			req._method = _MConnect
		}
	case 'O':
		if method == MethodOptions {
			req._method = _MOptions
		}
	case 'T':
		if method == MethodTrace {
			req._method = _MTrace
		}
	}
}

func (req *Request) parsePath() {
	if req.url.ok {
		return
	}
	rawPath := req.fl2
	u := &req.url
	u.ok = true

	portInd := bytes.IndexByte(rawPath, ':')
	if portInd >= 0 {
		u.Host = rawPath[:portInd]
		rawPath = rawPath[portInd+1:]
		pathInd := bytes.IndexByte(rawPath, '/')
		if pathInd >= 0 {
			u.Port = rawPath[:pathInd]
			rawPath = rawPath[pathInd:]
		} else {
			u.Port = rawPath
			return
		}
	}

	queryInd := bytes.IndexByte(rawPath, '?')
	if queryInd >= 0 {
		u.Query = rawPath[queryInd+1:]
		u.Path = rawPath[:queryInd]
	} else {
		u.Path = rawPath
	}
}
