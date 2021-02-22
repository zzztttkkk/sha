package sha

import (
	"bytes"
	"github.com/zzztttkkk/sha/utils"
	"os"
	"sync"
)

type Form struct {
	utils.Kvs
}

func (form *Form) onItem(k []byte, v []byte) { form.AppendBytes(decodeURIFormed(k), decodeURIFormed(v)) }

func (form *Form) FromUrlEncoded(p []byte) {
	var key []byte
	var val []byte
	var f bool
	for _, d := range p {
		if d == '&' {
			form.onItem(key, val)
			key = key[:0]
			val = val[:0]
			f = false
			continue
		}
		if d == '=' {
			f = true
			continue
		}
		if f {
			val = append(val, d)
		} else {
			key = append(key, d)
		}
	}
	form.onItem(key, val)
}

type FormFile struct {
	Name     string
	FileName string
	Header   Header

	buf []byte
}

func (file *FormFile) Data() []byte { return file.buf }

func (file *FormFile) Save(name string) error {
	f, e := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(file.buf)
	return e
}

type FormFiles []*FormFile

func (files FormFiles) Get(name []byte) *FormFile {
	for _, v := range files {
		if v.Name == string(name) {
			return v
		}
	}
	return nil
}

func (files FormFiles) GetAll(name []byte) []*FormFile {
	var rv []*FormFile
	for _, v := range files {
		if v.Name == string(name) {
			rv = append(rv, v)
		}
	}
	return rv
}

func (req *Request) parseQuery() {
	if !req.gotQuestionMark {
		req.questionMarkIndex = _QueryParsed
		return
	}
	req.query.FromUrlEncoded(req.RawPath[req.questionMarkIndex+1:])
	req.questionMarkIndex = _QueryParsed
}

func (req *Request) Query() *Form {
	if req.questionMarkIndex != _QueryParsed {
		req.parseQuery()
	}
	return &req.query
}

func (req *Request) QueryValue(name string) ([]byte, bool) {
	query := req.Query()
	if query != nil {
		return query.Get(name)
	}
	return nil, false
}

func (req *Request) QueryValues(name string) [][]byte {
	query := req.Query()
	if query != nil {
		return query.GetAll(name)
	}
	return nil
}

type _MultiPartParser struct {
	begin    []byte
	end      []byte
	line     []byte
	inHeader bool
	done     bool

	current *FormFile
	files   *FormFiles
	form    *Form
}

func (p *_MultiPartParser) reset() {
	p.begin = p.begin[:0]
	p.end = p.end[:0]
	p.line = p.line[:0]
	p.current = nil
	p.inHeader = false
	p.done = false
	p.files = nil
	p.form = nil
}

var multiPartParserPool = sync.Pool{New: func() interface{} { return &_MultiPartParser{} }}

func acquireMultiPartParser(request *Request) *_MultiPartParser {
	v := multiPartParserPool.Get().(*_MultiPartParser)
	v.files = &request.files
	v.form = &request.body
	return v
}

func releaseMultiPartParser(p *_MultiPartParser) {
	p.reset()
	multiPartParserPool.Put(p)
}

func (p *_MultiPartParser) setBoundary(boundary []byte) {
	p.begin = append(p.begin, '-', '-')
	p.begin = append(p.begin, boundary...)
	p.end = append(p.end, p.begin...)
	p.end = append(p.end, '-', '-')
	p.begin = append(p.begin, '\r', '\n')
	p.end = append(p.end, '\r', '\n')
}

func (p *_MultiPartParser) onHeaderLineOk() bool {
	ind := bytes.IndexByte(p.line, ':')
	if ind < 0 {
		return false
	}
	p.current.Header.AppendBytes(p.line[:ind], p.line[ind+2:])
	return true
}

func (p *_MultiPartParser) appendLine() {
	p.current.buf = append(p.current.buf, p.line...)
}

var formDataStr = []byte("form-data;")
var headerValueAttrsSep = []byte(";")
var nameStr = []byte("name=")
var filenameStr = []byte("filename=")

func (p *_MultiPartParser) onFieldOk() bool {
	disposition, ok := p.current.Header.Get(HeaderContentDisposition)
	if !ok || len(disposition) < 1 {
		return false
	}
	if !bytes.HasPrefix(disposition, formDataStr) {
		return false
	}

	var name []byte
	var filename []byte

	for _, v := range bytes.Split(disposition[11:], headerValueAttrsSep) {
		v = utils.InplaceTrimAsciiSpace(v)

		if bytes.HasPrefix(v, nameStr) {
			name = decodeURI(v[6 : len(v)-1])
			continue
		}
		if bytes.HasPrefix(v, filenameStr) {
			filename = decodeURI(v[10 : len(v)-1])
		}
	}

	p.current.buf = p.current.buf[:len(p.current.buf)-2] // \r\n

	if len(filename) > 0 {
		p.current.Name = utils.S(name)
		p.current.FileName = utils.S(filename)
		*p.files = append(*p.files, p.current)
		p.current = nil
	} else {
		p.form.AppendBytes(name, p.current.buf)
	}
	return true
}

func (req *Request) parseMultiPart(boundary []byte) bool {
	buf := *req.bodyBufferPtr
	parser := acquireMultiPartParser(req)
	defer releaseMultiPartParser(parser)

	parser.setBoundary(boundary)

	begin := false
	for _, b := range buf {
		if b == '\n' {
			parser.line = append(parser.line, b)
			if parser.inHeader { // is header line
				parser.line = utils.InplaceTrimAsciiSpace(parser.line)
				if len(parser.line) == 0 {
					parser.inHeader = false
				} else {
					if !parser.onHeaderLineOk() {
						return false
					}
				}
			} else {
				if bytes.Equal(parser.begin, parser.line) {
					if begin && !parser.onFieldOk() {
						return false
					}
					parser.inHeader = true
					begin = true
					if parser.current == nil {
						parser.current = &FormFile{}
					} else {
						parser.current.Header.Reset()
						parser.current.buf = parser.current.buf[:0]
					}
				} else if bytes.Equal(parser.end, parser.line) {
					if !parser.onFieldOk() {
						return false
					}
					parser.done = true
					break
				} else {
					parser.appendLine()
				}
			}
			parser.line = parser.line[:0]
			continue
		}
		parser.line = append(parser.line, b)
	}
	return true
}

var boundaryBytes = []byte("boundary=")

const (
	_BodyUnParsed = iota
	_BodyUnsupportedType
	_BodyOK
)

func (req *Request) parseBodyBuf() {
	if req.bodyBufferPtr == nil {
		req.bodyStatus = _BodyUnsupportedType
		return
	}

	buf := *req.bodyBufferPtr
	if len(buf) < 1 {
		req.bodyStatus = _BodyUnsupportedType
		return
	}

	typeValue := req.Header.ContentType()
	if len(typeValue) < 1 {
		req.bodyStatus = _BodyUnsupportedType
		return
	}

	if bytes.HasPrefix(typeValue, utils.B(MIMEForm)) {
		req.body.FromUrlEncoded(buf)
		req.bodyStatus = _BodyOK
		return
	}

	if bytes.HasPrefix(typeValue, utils.B(MIMEMultiPart)) {
		ind := bytes.Index(typeValue, boundaryBytes)
		if ind < 1 {
			req.bodyStatus = _BodyUnsupportedType
			return
		}

		req.parseMultiPart(typeValue[ind+9:])
		req.bodyStatus = _BodyOK
		return
	}
	req.bodyStatus = _BodyUnsupportedType
}

func (req *Request) BodyForm() *Form {
	if req.bodyStatus == _BodyUnParsed {
		req.parseBodyBuf()
	}
	if req.bodyStatus != _BodyOK {
		return nil
	}
	return &req.body
}

func (req *Request) BodyFormValue(name string) ([]byte, bool) {
	form := req.BodyForm()
	if form == nil {
		return nil, false
	}
	return form.Get(name)
}

func (req *Request) BodyFormValues(name string) [][]byte {
	form := req.BodyForm()
	if form == nil {
		return nil
	}
	return form.GetAll(name)
}

func (req *Request) Files() FormFiles {
	if req.bodyStatus == 0 {
		req.parseBodyBuf()
	}
	if req.bodyStatus != 2 {
		return nil
	}
	return req.files
}

func (req *Request) BodyRaw() []byte {
	if req.bodyBufferPtr == nil {
		return nil
	}
	return *req.bodyBufferPtr
}

// ctx
func (ctx *RequestCtx) FormValue(name string) ([]byte, bool) {
	v, ok := ctx.Request.QueryValue(name)
	if ok {
		return v, true
	}
	return ctx.Request.BodyFormValue(name)
}

func (ctx *RequestCtx) FormValues(name string) [][]byte {
	v := ctx.Request.QueryValues(name)
	v = append(v, ctx.Request.BodyFormValues(name)...)
	return v
}

func (ctx *RequestCtx) PathParam(name string) ([]byte, bool) {
	return ctx.Request.Params.Get(name)
}

func (ctx *RequestCtx) File(name []byte) *FormFile {
	return ctx.Request.Files().Get(name)
}

func (ctx *RequestCtx) Files(name []byte) []*FormFile {
	return ctx.Request.Files().GetAll(name)
}

func (ctx *RequestCtx) BodyRaw() []byte { return ctx.buf }
