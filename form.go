package sha

import (
	"bytes"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"mime"
	"os"
	"sync"
)

type Form struct {
	utils.Kvs
}

func (form *Form) onItem(k []byte, v []byte) {
	form.AppendBytes(utils.DecodeURIFormed(k), utils.DecodeURIFormed(v))
}

func (form *Form) FromURLEncoded(p []byte) {
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

func (form *Form) EncodeToBuf(w interface {
	io.ByteWriter
	io.Writer
}) {
	ind := 0
	qs := form.Size()
	form.EachItem(func(item *utils.KvItem) bool {
		ind++
		utils.EncodeURIComponentToBuf(item.Key, w)
		_ = w.WriteByte('=')
		utils.EncodeURIComponentToBuf(item.Val, w)
		if ind < qs {
			_ = w.WriteByte('&')
		}
		return true
	})
}

type FormFile struct {
	Name     string
	FileName string
	Header   Header

	buf    []byte
	cursor int64
}

func (file *FormFile) meta() {
	disposition, _ := file.Header.Get(HeaderContentDisposition)
	mt, v, e := mime.ParseMediaType(utils.S(disposition))
	if e != nil {
		return
	}
	if mt != "form-data" {
		return
	}
	file.Name = v["name"]
	file.FileName = v["filename"]
}

func (file *FormFile) reset(maxCap int) {
	file.Name = ""
	file.FileName = ""
	file.Header.Reset()
	file.buf = file.buf[:0]
	if maxCap > 0 && cap(file.buf) > maxCap {
		file.buf = nil
	}
	file.cursor = 0
}

func (file *FormFile) Read(p []byte) (int, error) {
	var (
		fileSize = int64(len(file.buf))
		bufSize  = int64(len(p))
	)

	if file.cursor >= fileSize {
		return 0, io.EOF
	}

	if file.cursor+bufSize <= fileSize {
		file.cursor += bufSize
		copy(p, file.buf[file.cursor:file.cursor+bufSize-1])
		return len(p), nil
	}
	file.cursor = fileSize
	copy(p, file.buf[file.cursor:])
	return int(fileSize - file.cursor), nil
}

func (file *FormFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		file.cursor += offset
	case io.SeekStart:
		file.cursor = offset
	case io.SeekEnd:
		file.cursor = int64(len(file.buf)) - offset
	}
	if file.cursor < 0 {
		file.cursor = 0
	}
	return file.cursor, nil
}

var _ io.ReadSeeker = (*FormFile)(nil)

var formFilePool = sync.Pool{New: func() interface{} { return &FormFile{} }}

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

func (files FormFiles) Get(name string) *FormFile {
	for _, v := range files {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (files FormFiles) GetAll(name string) []*FormFile {
	var rv []*FormFile
	for _, v := range files {
		if v.Name == name {
			rv = append(rv, v)
		}
	}
	return rv
}

func (req *Request) parseQuery() {
	if req.queryParsed {
		return
	}
	req.queryParsed = true
	if len(req.URL.Query) > 0 {
		req.query.FromURLEncoded(req.URL.Query)
	}
}

func (req *Request) Query() *Form {
	req.parseQuery()
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

func (req *Request) parseMultiPartForm() bool {
	var (
		status  int
		begin   int
		current *FormFile
		hItem   *utils.KvItem
		keyDone bool
		skipSp  bool
		skipNL  bool
		end     bool
		data    = req.body.Bytes()
	)

	for ind, b := range data {
		switch status {
		case 0:
			req.boundaryLine = append(req.boundaryLine, b)
			if b == '\n' {
				if !bytes.Equal(req.boundaryLine, req.boundaryBegin) {
					return false
				}
				status++
				req.boundaryLine = req.boundaryLine[:0]
				continue
			}
			if len(req.boundaryLine) > len(req.boundaryBegin) {
				return false
			}
			continue
		case 1:
			if b == ':' {
				keyDone = true
				skipSp = true
				continue
			}

			if skipSp {
				if b == ' ' {
					skipSp = false
					continue
				}
				return false
			}

			if skipNL {
				if b == '\n' {
					skipNL = false
					continue
				}
				return false
			}

			if b == '\r' {
				keyDone = false
				skipNL = true
				if hItem == nil {
					status++
					begin = ind + 2
					continue
				}
				hItem = nil
				continue
			}

			if hItem == nil {
				if current == nil {
					current = formFilePool.Get().(*FormFile)
					current.Header.fromOutSide = true
				}

				hItem = current.Header.AppendBytes(nil, nil)
			}

			if keyDone {
				hItem.Val = append(hItem.Val, b)
			} else {
				if b > 127 {
					return false
				}
				hItem.Key = append(hItem.Key, toLowerTable[b])
			}
		case 2:
			if skipNL {
				if b == '\n' {
					skipNL = false
					continue
				}
				return false
			}

			if b == '\n' {
				req.boundaryLine = append(req.boundaryLine, b)

				end = bytes.Equal(req.boundaryLine, req.boundaryEnd)
				if end || bytes.Equal(req.boundaryLine, req.boundaryBegin) {
					if current == nil {
						return false
					}
					current.buf = data[begin : ind-len(req.boundaryLine)-1]
					current.meta()

					if len(current.Name) < 1 {
						current.reset(0)
						formFilePool.Put(current)
					} else {
						if len(current.FileName) > 0 {
							req.files = append(req.files, current)
						} else {
							item := req.bodyForm.AppendBytes(nil, nil)
							item.Key = append(item.Key, current.Name...)
							item.Val = append(item.Val, current.buf...)
							current.reset(0)
							formFilePool.Put(current)
						}
					}

					status = 1
					current = nil
				}

				if end {
					return true
				}
				req.boundaryLine = req.boundaryLine[:0]
				continue
			}
			if len(req.boundaryLine) >= len(req.boundaryEnd) {
				continue
			}
			req.boundaryLine = append(req.boundaryLine, b)
		}
	}
	return false
}

var boundaryBytes = []byte("boundary=")

const (
	_BodyUnParsed = iota
	_BodyUnsupportedType
	_BodyOK
)

func (req *Request) parseBodyBuf() {
	buf := req._HTTPPocket.body
	if buf == nil || buf.Len() < 1 {
		req.bodyStatus = _BodyUnsupportedType
		return
	}

	typeValue := req.Header().ContentType()
	if len(typeValue) < 1 {
		req.bodyStatus = _BodyUnsupportedType
		return
	}

	if bytes.HasPrefix(typeValue, utils.B(MIMEForm)) {
		req.bodyForm.FromURLEncoded(buf.Bytes())
		req.bodyStatus = _BodyOK
		return
	}

	if bytes.HasPrefix(typeValue, utils.B(MIMEMultiPart)) {
		ind := bytes.Index(typeValue, boundaryBytes)
		if ind < 1 {
			req.bodyStatus = _BodyUnsupportedType
			return
		}

		req.boundaryBegin = append(req.boundaryBegin, "--"...)
		req.boundaryBegin = append(req.boundaryBegin, typeValue[ind+9:]...)

		req.boundaryEnd = append(req.boundaryEnd, req.boundaryBegin...)
		req.boundaryEnd = append(req.boundaryEnd, "--\r\n"...)
		req.boundaryBegin = append(req.boundaryBegin, "\r\n"...)
		if req.parseMultiPartForm() {
			req.bodyStatus = _BodyOK
		} else {
			req.bodyStatus = _BodyUnsupportedType
		}
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
	return &req.bodyForm
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
	if req.bodyStatus == _BodyUnParsed {
		req.parseBodyBuf()
	}
	if req.bodyStatus != _BodyOK {
		return nil
	}
	return req.files
}

func (req *Request) BodyRaw() []byte {
	if req._HTTPPocket.body == nil {
		return nil
	}
	return req._HTTPPocket.body.Bytes()
}

func (req *Request) FormValue(name string) ([]byte, bool) {
	v, ok := req.BodyFormValue(name)
	if ok {
		return v, ok
	}
	return req.QueryValue(name)
}

func (req *Request) FormValues(name string) [][]byte {
	v := req.BodyFormValues(name)
	return append(v, req.QueryValues(name)...)
}
