package sha

import (
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"time"
)

type CookieSameSiteVal string

const (
	CookeSameSiteDefault = CookieSameSiteVal("")
	CookieSameSiteLax    = CookieSameSiteVal("lax")
	CookieSameSiteStrict = CookieSameSiteVal("strict")
	CookieSameSizeNone   = CookieSameSiteVal("none")
)

type CookieOptions struct {
	Domain   string
	Path     string
	MaxAge   int64
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
	SameSite CookieSameSiteVal
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
	item := res.header.Append(HeaderSetCookie, nil)

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

func (req *Request) Cookies() *utils.Kvs { return &req.cookies }

func (req *Request) CookieValue(key string) ([]byte, bool) {
	if !req.cookieParsed {
		v, ok := req.Header().Get(HeaderCookie)
		if ok {
			var item *utils.KvItem
			var keyDone bool
			var skipSp bool

			for _, b := range v {
				switch b {
				case '=':
					keyDone = true
					continue
				case ';':
					item = nil
					skipSp = true
					keyDone = false
					continue
				case ' ':
					if skipSp {
						skipSp = false
						continue
					}
					goto appendByte
				default:
					goto appendByte
				}

			appendByte:
				if item == nil {
					item = req.cookies.AppendBytes(nil, nil)
				}
				if keyDone {
					item.Val = append(item.Val, b)
				} else {
					item.Key = append(item.Key, b)
				}
			}
		}
		req.cookieParsed = true
	}
	return req.cookies.Get(key)
}
