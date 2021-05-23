package sha

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
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
	ok     bool
	Host   []byte
	Port   []byte
	Path   []byte
	Query  []byte
	Params URLParams
}

func (u *URL) String() string {
	return fmt.Sprintf("Host: %s, Port: %s, Path: %s, Query: %s", u.Host, u.Port, u.Path, u.Query)
}

func (u *URL) reset() {
	u.ok = false
	u.Host = nil
	u.Port = nil
	u.Path = nil
	u.Query = nil
	u.Params.Reset()
}

type Request struct {
	noCopy
	_HTTPPocket

	_method _Method

	URL         URL
	queryParsed bool

	// forms
	query      Form
	bodyStatus int // 0: unparsed; 1: unsupported content type; 2: parsed
	bodyForm   Form
	// forms ---  multipart
	files         FormFiles
	boundaryBegin []byte
	boundaryEnd   []byte
	boundaryLine  []byte

	cookies      utils.Kvs
	cookieParsed bool

	history []string // redirect history
}

func (req *Request) Reset() {
	req._HTTPPocket.reset()
	req._method = 0

	req.URL.reset()

	req.queryParsed = false
	req.query.Reset()

	req.bodyForm.Reset()
	req.bodyStatus = _BodyUnParsed

	for _, f := range req.files {
		f.reset()
		formFilePool.Put(f)
	}
	req.files = req.files[:0]
	req.boundaryBegin = req.boundaryBegin[:0]
	req.boundaryEnd = req.boundaryEnd[:0]
	req.boundaryLine = req.boundaryLine[:0]

	req.cookies.Reset()
	req.cookieParsed = false
	req.history = nil
}

func (req *Request) History() []string { return req.history }

func (req *Request) Method() []byte { return req.fl1 }

func (req *Request) RawPath() []byte { return req.fl2 }

func (req *Request) Path() string {
	req.parsePath()
	if len(req.URL.Path) > 0 {
		return utils.S(req.URL.Path)
	}
	return "/"
}

func (req *Request) HTTPVersion() []byte { return req.fl3 }

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
	return req.SetPathString(utils.S(path))
}

func (req *Request) SetPathString(path string) *Request {
	req.fl2 = req.fl2[:0]
	req.fl2 = append(req.fl2, path...)
	req.URL.ok = false
	return req
}

func (req *Request) SetJSONBody(v interface{}) error {
	req.Header().SetContentType(MIMEJson)
	return jsonx.NewEncoder(&req._HTTPPocket).Encode(v)
}

func (req *Request) SetMultiValueMapBody(fv utils.MultiValueMap) *Request {
	req.header.SetContentType(MIMEForm)
	req.bodyForm.LoadMap(fv)
	return req
}

func (req *Request) SetFormBody(form *Form) *Request {
	req.header.SetContentType(MIMEForm)
	req.bodyForm.LoadKvs(&form.Kvs)
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
			return
		}
	case 'H':
		if method == MethodHead {
			req._method = _MHead
			return
		}
	case 'P':
		switch method {
		case MethodPatch:
			req._method = _MPatch
			return
		case MethodPost:
			req._method = _MPost
			return
		case MethodPut:
			req._method = _MPut
			return
		}
	case 'D':
		if method == MethodDelete {
			req._method = _MDelete
			return
		}
	case 'C':
		if method == MethodConnect {
			req._method = _MConnect
			return
		}
	case 'O':
		if method == MethodOptions {
			req._method = _MOptions
			return
		}
	case 'T':
		if method == MethodTrace {
			req._method = _MTrace
			return
		}
	}
	req._method = _MCustom
}

func (req *Request) parsePath() {
	if req.URL.ok {
		return
	}
	rawPath := req.RawPath()
	u := &req.URL
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
