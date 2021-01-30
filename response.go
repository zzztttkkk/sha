package sha

import (
	"bufio"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"sync"
	"time"
)

type Response struct {
	statusCode int
	Header     Header

	headerBuf          []byte
	sendBuf            *bufio.Writer
	bodyBuf            *internal.Buf
	compressWriter     _CompressionWriter
	compressWriterPool *sync.Pool
}

func (res *Response) Write(p []byte) (int, error) {
	if res.compressWriter != nil {
		return res.compressWriter.Write(p)
	}
	res.bodyBuf.Data = append(res.bodyBuf.Data, p...)
	return len(p), nil
}

func (res *Response) SetStatusCode(v int) {
	res.statusCode = v
}

func (res *Response) ResetBodyBuffer() {
	res.bodyBuf.Data = res.bodyBuf.Data[:0]
	if res.compressWriter != nil {
		res.compressWriter.Reset(res.bodyBuf)
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

const (
	_CookieSep      = "; "
	_CookieDomain   = "Domain="
	_CookiePath     = "Path="
	_CookieExpires  = "Expires="
	_CookieMaxAge   = "Max-Age="
	_CookieSecure   = "Secure"
	_CookieHttpOnly = "Httponly"
	_CookieSameSite = "Samesite="
)

var defaultCookieOptions CookieOptions

func (res *Response) SetCookie(k, v string, options *CookieOptions) {
	if options == nil {
		options = &defaultCookieOptions
	}
	item := res.Header.Append(HeaderSetCookie, nil)

	item.Val = append(item.Val, utils.B(k)...)
	item.Val = append(item.Val, '=')
	item.Val = append(item.Val, utils.B(v)...)
	item.Val = append(item.Val, _CookieSep...)

	if len(options.Domain) > 0 {
		item.Val = append(item.Val, _CookieDomain...)
		item.Val = append(item.Val, utils.B(options.Domain)...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if len(options.Path) > 0 {
		item.Val = append(item.Val, _CookiePath...)
		item.Val = append(item.Val, utils.B(options.Path)...)
		item.Val = append(item.Val, _CookieSep...)
	} else {
		item.Val = append(item.Val, _CookiePath...)
		item.Val = append(item.Val, '/')
		item.Val = append(item.Val, _CookieSep...)
	}

	if !options.Expires.IsZero() {
		item.Val = append(item.Val, _CookieExpires...)
		item.Val = append(item.Val, utils.B(options.Expires.Format(time.RFC1123))...)
		item.Val = append(item.Val, _CookieSep...)
	} else if options.MaxAge > 0 {
		item.Val = append(item.Val, _CookieMaxAge...)
		item.Val = append(item.Val, utils.B(strconv.FormatInt(options.MaxAge, 10))...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if options.Secure {
		item.Val = append(item.Val, _CookieSecure...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if options.HttpOnly {
		item.Val = append(item.Val, _CookieHttpOnly...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if len(options.SameSite) > 0 {
		item.Val = append(item.Val, _CookieSameSite...)
	}
}

func (res *Response) reset() {
	res.statusCode = 0
	res.headerBuf = res.headerBuf[:0]
	res.Header.Reset()
	res.ResetBodyBuffer()
}
