package sha

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"sync"
	"time"
)

type Response struct {
	_HTTPPocket
	statusCode int
	cw         _CompressionWriter
	cwp        *sync.Pool
}

func (res *Response) StatusCode() int { return res.statusCode }

func (res *Response) Phrase() string { return utils.S(res.fl3) }

func (res *Response) HTTPVersion() string { return utils.S(res.fl1) }

func (res *Response) SetHTTPVersion(v string) *Response {
	res.fl1 = res.fl1[:0]
	res.fl1 = append(res.fl1, v...)
	return res
}

var ErrUnknownResponseStatusCode = fmt.Errorf("sha: unknown response status code")

func (res *Response) SetStatusCode(v int) *Response {
	res.statusCode = v
	res.fl3 = statusTextMap[v]
	if len(res.fl3) < 1 {
		panic(ErrUnknownResponseStatusCode)
	}
	res.fl2 = append(res.fl2, strconv.FormatInt(int64(v), 10)...)
	return res
}

func (res *Response) Body() *bytes.Buffer { return res.body }

func (res *Response) Write(p []byte) (int, error) {
	if res.cw != nil {
		return res.cw.Write(p)
	}
	return res.body.Write(p)
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
	item := res.Header().Append(HeaderSetCookie, nil)

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
	res._HTTPPocket.reset()
	res.parseStatus = 0
	res.statusCode = 0
	if res.cw != nil {
		res.cw.Reset(nil)
		res.cwp.Put(res.cw)
		res.cw = nil
		res.cwp = nil
	}
}

func (res *Response) ResetBody() {
	if res.body != nil {
		res.body.Reset()
	}
	if res.cw != nil {
		res.cw.Reset(&res._HTTPPocket)
	}
}
