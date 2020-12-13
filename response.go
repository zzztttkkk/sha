package sha

import (
	"github.com/zzztttkkk/sha/internal"
	"strconv"
	"sync"
	"time"
)

type Response struct {
	statusCode int
	Header     Header

	buf            *internal.Buf
	compressWriter _CompressionWriter
	cwrPool        *sync.Pool

	headerWritten bool
}

func (res *Response) Write(p []byte) (int, error) {
	if res.compressWriter != nil {
		return res.compressWriter.Write(p)
	}
	res.buf.Data = append(res.buf.Data, p...)
	return len(p), nil
}

func (res *Response) SetStatusCode(v int) {
	res.statusCode = v
}

func (res *Response) ResetBodyBuffer() {
	res.buf.Data = res.buf.Data[:0]
	if res.compressWriter != nil {
		res.compressWriter.Reset(res.buf)
	}
}

type _SameSiteVal string

const (
	CookeSameSiteDefault = _SameSiteVal("")
	CookieSameSiteLax    = _SameSiteVal("lax")
	CookieSameSiteStrict = _SameSiteVal("strict")
	CookieSameSizeNone   = _SameSiteVal("none")
)

type CookieOptions struct {
	Domain   string
	Path     string
	MaxAge   int64
	Expires  time.Time
	Secure   bool
	HttpOnly bool
	SameSite _SameSiteVal
}

func (res *Response) SetCookie(k, v string, options CookieOptions) {
	item := res.Header.Append(internal.B(HeaderSetCookie), nil)

	item.Val = append(item.Val, internal.B(k)...)
	item.Val = append(item.Val, '=')
	item.Val = append(item.Val, internal.B(v)...)
	item.Val = append(item.Val, ';')
	item.Val = append(item.Val, ' ')

	if len(options.Domain) > 0 {
		item.Val = append(item.Val, 'D', 'o', 'm', 'a', 'i', 'n', '=')
		item.Val = append(item.Val, internal.B(options.Domain)...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if len(options.Path) > 0 {
		item.Val = append(item.Val, 'P', 'a', 't', 'h', '=')
		item.Val = append(item.Val, internal.B(options.Path)...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if !options.Expires.IsZero() {
		item.Val = append(item.Val, 'E', 'x', 'p', 'i', 'r', 'e', 's', '=')
		item.Val = append(item.Val, internal.B(options.Expires.Format(time.RFC1123))...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	} else {
		item.Val = append(item.Val, 'M', 'a', 'x', '-', 'A', 'g', 'e', '=')
		item.Val = append(item.Val, internal.B(strconv.FormatInt(options.MaxAge, 10))...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if options.Secure {
		item.Val = append(item.Val, 'S', 'e', 'c', 'u', 'r', 'e')
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if options.HttpOnly {
		item.Val = append(item.Val, 'H', 't', 't', 'p', 'o', 'n', 'l', 'y')
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if len(options.SameSite) > 0 {
		item.Val = append(item.Val, 'S', 'a', 'm', 'e', 's', 'i', 't', 'e', '=')
		item.Val = append(item.Val, internal.B(string(options.SameSite))...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}
}

func (res *Response) reset() {
	res.statusCode = 0
	res.Header.Reset()
	res.headerWritten = false

	res.ResetBodyBuffer()
}
