package sha

import (
	"bufio"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"sync"
	"time"
)

type Response struct {
	version    []byte
	statusCode int
	phrase     []byte

	Header Header

	headerBuf          []byte
	sendBuf            *bufio.Writer
	bodyBuf            *utils.Buf
	compressWriter     _CompressionWriter
	compressWriterPool *sync.Pool
	parseStatus        int
}

func (res *Response) setVersion(v []byte) {
	if res.version == nil {
		res.version = make([]byte, 0, len(v))
	} else {
		res.version = res.phrase[:0]
	}
	res.version = append(res.version, v...)
}

func (res *Response) setPhrase(v []byte) {
	if res.phrase == nil {
		res.phrase = make([]byte, 0, len(v))
	} else {
		res.phrase = res.phrase[:0]
	}
	res.phrase = append(res.phrase, v...)
}

func (res *Response) StatusCode() int { return res.statusCode }

func (res *Response) Phrase() string { return utils.S(res.phrase) }

func (res *Response) ProtocolVersion() string { return utils.S(res.version) }

func (res *Response) String() string {
	return fmt.Sprintf("sha.Response<%d, %s>", res.statusCode, res.phrase)
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
	HTTPOnly bool
	SameSite _SameSiteVal
}

const (
	_CookieSep      = "; "
	_CookieDomain   = "Domain="
	_CookiePath     = "Path="
	_CookieExpires  = "Expires="
	_CookieMaxAge   = "Max-Age="
	_CookieSecure   = "Secure"
	_CookieHTTPOnly = "HTTPOnly"
	_CookieSameSite = "Samesite="
)

var defaultCookieOptions CookieOptions

func (res *Response) SetCookie(k, v string, options *CookieOptions) {
	if options == nil {
		options = &defaultCookieOptions
	}
	item := res.Header.Append(HeaderSetCookie, nil)

	item.Val = append(item.Val, k...)
	item.Val = append(item.Val, '=')
	item.Val = append(item.Val, v...)
	item.Val = append(item.Val, _CookieSep...)

	if len(options.Domain) > 0 {
		item.Val = append(item.Val, _CookieDomain...)
		item.Val = append(item.Val, options.Domain...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if len(options.Path) > 0 {
		item.Val = append(item.Val, _CookiePath...)
		item.Val = append(item.Val, options.Path...)
		item.Val = append(item.Val, _CookieSep...)
	} else {
		item.Val = append(item.Val, _CookiePath...)
		item.Val = append(item.Val, '/')
		item.Val = append(item.Val, _CookieSep...)
	}

	if !options.Expires.IsZero() {
		item.Val = append(item.Val, _CookieExpires...)
		item.Val = append(item.Val, options.Expires.Format(time.RFC1123)...)
		item.Val = append(item.Val, _CookieSep...)
	} else {
		item.Val = append(item.Val, _CookieMaxAge...)
		item.Val = append(item.Val, strconv.FormatInt(options.MaxAge, 10)...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if options.HTTPOnly {
		item.Val = append(item.Val, _CookieHTTPOnly...)
		item.Val = append(item.Val, _CookieSep...)
	}

	if len(options.SameSite) > 0 {
		item.Val = append(item.Val, _CookieSameSite...)
		item.Val = append(item.Val, options.SameSite...)

		if options.SameSite == CookieSameSizeNone {
			options.Secure = true
		}
	}

	if options.Secure {
		item.Val = append(item.Val, _CookieSecure...)
		item.Val = append(item.Val, _CookieSep...)
	}
}

func (res *Response) reset() {
	res.statusCode = 0
	res.phrase = nil
	res.version = nil
	res.headerBuf = res.headerBuf[:0]
	res.Header.Reset()
	res.parseStatus = 0
	res.ResetBodyBuffer()
}

func (res *Response) Body() []byte { return res.bodyBuf.Data }
