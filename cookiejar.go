package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/jsonx"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

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

type CookieJar struct {
	all map[string]map[string]*CookieItem
}

func (cm *CookieJar) append(host, name string, opt *CookieItem) {
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

	return false
}

func quoted(a, b byte) bool { return a == b && (a == '"' || a == '\'') }

func (cm *CookieJar) Update(host, item string) error {
	var key string
	var obj = &CookieItem{Created: time.Now()}
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
			default:
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
			continue
		}

		_key := strings.ToLower(strings.TrimSpace(kv[:ind]))
		_val := strings.TrimSpace(kv[ind+1:])
		if len(key) < 1 {
			key = _key
			if quoted(_val[0], _val[len(_val)-1]) {
				_val = _val[1 : len(_val)-1]
			}
			obj.Value = _val
			continue
		}

		switch _key {
		case "domain":
			if !checkDomain(_val, host) {
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
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

func (cm *CookieJar) Load(reader io.Reader) error {
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

func NewCookieJar() *CookieJar { return &CookieJar{all: map[string]map[string]*CookieItem{}} }
