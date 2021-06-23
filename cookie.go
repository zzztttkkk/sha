package sha

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
)

type CookieSameSiteVal string

const (
	CookeSameSiteDefault = CookieSameSiteVal("")
	CookieSameSiteLax    = CookieSameSiteVal("lax")
	CookieSameSiteStrict = CookieSameSiteVal("strict")
	CookieSameSizeNone   = CookieSameSiteVal("none")
)

type CookieOptions struct {
	Domain   string            `json:"domain"`
	Path     string            `json:"path"`
	MaxAge   int64             `json:"maxage"`
	Expires  time.Time         `json:"expires"`
	Secure   bool              `json:"secure"`
	HTTPOnly bool              `json:"http_only"`
	SameSite CookieSameSiteVal `json:"same_site"`
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

type CookieItem struct {
	CookieOptions
	Created time.Time `json:"created"`
	Value   string    `json:"value"`
}

func (item *CookieItem) isValid() bool {
	if item.MaxAge > 0 {
		return int64(time.Since(item.Created)/time.Second) <= item.MaxAge
	}
	if item.Expires.IsZero() {
		return false
	}
	return time.Since(item.Expires) <= 0
}

type _CookieManager struct {
	all map[string]map[string]*CookieItem
}

func (cm *_CookieManager) append(host, name string, opt *CookieItem) {
	host = strings.ToLower(host)
	if cm.all == nil {
		cm.all = map[string]map[string]*CookieItem{}
	}
	hm := cm.all[host]
	if hm == nil {
		hm = map[string]*CookieItem{}
		cm.all[host] = hm
	}
	if opt.isValid() {
		hm[name] = opt
		return
	}
	delete(hm, name)
}

func checkDomain(domain, host string) bool {
	if domain == host {
		return true
	}

	if len(domain) < 5 { // *.a.b
		return false
	}
	if !strings.HasPrefix(domain, "*.") {
		return false
	}

	return true
}

func (cm *_CookieManager) Update(host, item string) error {
	var key string
	var obj = &CookieItem{}
	for _, kv := range strings.Split(item, ";") {
		kv = strings.TrimSpace(kv)
		if len(kv) < 1 {
			continue
		}
		ind := strings.IndexRune(kv, '=')
		if ind < 0 {
			if len(key) < 1 {
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
			switch strings.ToLower(kv) {
			case "secure":
				obj.Secure = true
			case "httponly":
				obj.HTTPOnly = true
			case "samesite":
				break
			default:
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
		}

		_key := strings.ToLower(strings.TrimSpace(kv[:ind]))
		_val := strings.TrimSpace(kv[ind+1:])
		if len(key) < 1 {
			key = _key
			obj.Value = _val
			continue
		}

		switch _key {
		case "domain":
			// todo domain must match host
			obj.Domain = _val
		case "path":
			obj.Path = _val
		case "maxage":
			v, e := strconv.ParseInt(_val, 10, 64)
			if e != nil {
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
			obj.MaxAge = v
		case "expires":
			t, e := time.Parse(time.RFC1123, _val)
			if e != nil {
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
			obj.Expires = t
		case "samesite":
			switch _val {
			case "":
				break
			case "lax":
				obj.SameSite = CookieSameSiteLax
			case "strict":
				obj.SameSite = CookieSameSiteStrict
			case "none":
				obj.SameSite = CookieSameSizeNone
			default:
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
		default:
			return fmt.Errorf("sha: bad cookie value `%s`", item)
		}
	}

	cm.append(host, key, obj)
	return nil
}

func (cm *_CookieManager) Load(reader io.Reader) error {
	allBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var m map[string]map[string]*CookieItem
	if err = jsonx.Unmarshal(allBytes, &m); err != nil {
		return err
	}
	for k, v := range m {
		for name, opt := range v {
			cm.append(k, name, opt)
		}
	}
	return nil
}
