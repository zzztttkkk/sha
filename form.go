package suna

import (
	"bytes"
	"github.com/zzztttkkk/suna/internal"
	"io"
	"os"
	"sync"
)

type Form struct {
	internal.Kvs
}

func (form *Form) onItem(k []byte, v []byte) {
	form.Append(decodeURIFormed(k), decodeURIFormed(v))
}

func (form *Form) ParseUrlEncoded(p []byte) {
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

	buf    []byte
	cursor int
}

func (file *FormFile) Size() int {
	return len(file.buf)
}

func (file *FormFile) Seek(offset int64, whence int) (int64, error) {
	c := int64(file.cursor)
	l := int64(len(file.buf))
	switch whence {
	case io.SeekCurrent:
		c += offset
	case io.SeekEnd:
		c = l - 1 - offset
	case io.SeekStart:
		c = offset
	}
	if c > l-1 || c < 0 {
		return -1, io.ErrUnexpectedEOF
	}
	file.cursor = int(c)
	return c, nil
}

func (file *FormFile) Write(p []byte) (int, error) {
	file.buf = append(file.buf, p...)
	return len(p), nil
}

func (file *FormFile) Read(p []byte) (int, error) {
	s := 0
	for ; s < len(p) && file.cursor < len(file.buf); s++ {
		p[s] = file.buf[file.cursor+s]
	}
	if file.cursor == len(file.buf)-1 {
		return s, io.EOF
	}
	return s, nil
}

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
	req.query.ParseUrlEncoded(req.rawPath[req.queryStatus-1:])
	req.queryStatus = 1
}

func (req *Request) Query() *Form {
	if req.queryStatus > 2 {
		req.parseQuery()
	}
	if req.queryStatus != 1 {
		return nil
	}
	return &req.query
}

func (req *Request) QueryValue(name []byte) ([]byte, bool) {
	query := req.Query()
	if query != nil {
		return query.Get(name)
	}
	return nil, false
}

func (req *Request) QueryValues(name []byte) [][]byte {
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
	ind := bytes.Index(p.line, headerKVSep)
	if ind < 0 {
		return false
	}
	p.current.Header.Append(p.line[:ind], p.line[ind+2:])
	return true
}

func (p *_MultiPartParser) appendLine() {
	p.current.buf = append(p.current.buf, p.line...)
}

var formdataStr = []byte("form-data;")
var headerValueAttrsSep = []byte(";")
var nameStr = []byte("name=")
var filenameStr = []byte("filename=")

func (p *_MultiPartParser) onFieldOk() bool {
	disposition, ok := p.current.Header.Get(internal.B(HeaderContentDisposition))
	if !ok || len(disposition) < 1 {
		return false
	}
	if !bytes.HasPrefix(disposition, formdataStr) {
		return false
	}

	var name []byte
	var filename []byte

	for _, v := range bytes.Split(disposition[11:], headerValueAttrsSep) {
		v = internal.InplaceTrimAsciiSpace(v)

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
		p.current.Name = string(name)
		p.current.FileName = string(filename)
		*p.files = append(*p.files, p.current)
		p.current = nil
	} else {
		p.form.Append(name, p.current.buf)
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
				parser.line = internal.InplaceTrimAsciiSpace(parser.line)
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

func (req *Request) parseBodyBuf() {
	if req.bodyBufferPtr == nil {
		req.bodyStatus = 1
		return
	}

	buf := *req.bodyBufferPtr
	if len(buf) < 1 {
		req.bodyStatus = 1
		return
	}

	typeValue := req.Header.ContentType()
	if len(typeValue) < 1 {
		req.bodyStatus = 1
		return
	}

	if bytes.HasPrefix(typeValue, MIMEForm) {
		req.body.ParseUrlEncoded(buf)
		req.bodyStatus = 2
		return
	}

	if bytes.HasPrefix(typeValue, MIMEMultiPart) {
		ind := bytes.Index(typeValue, boundaryBytes)
		if ind < 1 {
			req.bodyStatus = 1
			return
		}

		req.parseMultiPart(typeValue[ind+9:])
		req.bodyStatus = 2
		return
	}
	req.bodyStatus = 1
}

func (req *Request) BodyForm() *Form {
	if req.bodyStatus == 0 {
		req.parseBodyBuf()
	}
	if req.bodyStatus != 2 {
		return nil
	}
	return &req.body
}

func (req *Request) BodyFormValue(name []byte) ([]byte, bool) {
	form := req.BodyForm()
	if form == nil {
		return nil, false
	}
	return form.Get(name)
}

func (req *Request) BodyFormValues(name []byte) [][]byte {
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

// ctx
func (ctx *RequestCtx) FormValue(name []byte) ([]byte, bool) {
	v, ok := ctx.Request.QueryValue(name)
	if ok {
		return v, true
	}
	return ctx.Request.BodyFormValue(name)
}

func (ctx *RequestCtx) FormValues(name []byte) [][]byte {
	v := ctx.Request.QueryValues(name)
	v = append(v, ctx.Request.BodyFormValues(name)...)
	return v
}

func (ctx *RequestCtx) PathParam(name []byte) ([]byte, bool) {
	return ctx.Request.Params.Get(name)
}

func (ctx *RequestCtx) File(name []byte) *FormFile {
	return ctx.Request.Files().Get(name)
}

func (ctx *RequestCtx) Files(name []byte) []*FormFile {
	return ctx.Request.Files().GetAll(name)
}
