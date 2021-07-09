package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/jsonx"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type CookieItem struct {
	CookieOptions
	Created time.Time `json:"created"`
	Value   string    `json:"value"`
	Key     string    `json:"key"`
}

func (item *CookieItem) isValid() bool {
	if item.MaxAge == 0 && item.Expires.IsZero() {
		return true
	}
	if item.MaxAge > 0 {
		return int64(time.Since(item.Created)/time.Second) <= item.MaxAge
	}
	if item.MaxAge < 0 {
		return false
	}
	return time.Since(item.Expires) <= 0
}

func (item *CookieItem) String() string {
	return fmt.Sprintf("CookieItem(%s=%s, domain: %s)", item.Key, item.Value, item.Domain)
}

type CookieJar struct {
	all map[string]map[string]*CookieItem
}

func (jar *CookieJar) append(opt *CookieItem) {
	domain := opt.Domain
	key := opt.Key

	if domain[0] != '.' {
		domain = "." + domain
	}
	if jar.all == nil {
		jar.all = map[string]map[string]*CookieItem{}
	}
	dm := jar.all[domain]
	if dm == nil {
		dm = map[string]*CookieItem{}
		jar.all[domain] = dm
	}
	if opt.isValid() {
		dm[key] = opt
		return
	}
	delete(dm, key)
}

// a.b
func isDomain(v string) bool {
	return len(v) > 2 && strings.ContainsRune(v, '.') && !strings.Contains(v, "..")
}

func checkDomain(domain, host string) bool {
	// host is a ip or localhost name
	if net.ParseIP(host) != nil || !isDomain(host) {
		return domain == host
	}
	return isDomain(domain) && strings.HasSuffix(domain, host)
}

func quoted(a, b byte) bool { return a == b && (a == '"' || a == '\'') }

func removePortAndLowercase(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx > -1 {
		s = s[:idx]
	}
	return strings.ToLower(s)
}

func (jar *CookieJar) Update(host, item string) error {
	host = removePortAndLowercase(host)

	var obj = &CookieItem{Created: time.Now()}
	for _, kv := range strings.Split(item, ";") {
		kv = strings.TrimSpace(kv)
		if len(kv) < 1 {
			continue
		}
		ind := strings.IndexRune(kv, '=')
		if ind < 0 {
			if len(obj.Key) < 1 {
				return fmt.Errorf("sha: bad cookie value `%s`", item)
			}
			kv = strings.TrimSpace(kv)
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
		if len(obj.Key) < 1 {
			obj.Key = _key
			if quoted(_val[0], _val[len(_val)-1]) {
				_val = _val[1 : len(_val)-1]
			}
			obj.Value = _val
			continue
		}

		switch _key {
		case "domain":
			_val = strings.ToLower(_val)
			if !checkDomain(_val, host) {
				return fmt.Errorf("sha: bad cookie value `%s`, domain not match host", item)
			}
			obj.Domain = _val
		case "path":
			obj.Path = _val
		case "maxage":
			v, e := strconv.ParseInt(_val, 10, 64)
			if e != nil {
				return fmt.Errorf("sha: bad cookie value `%s`, maxage is not a int", item)
			}
			obj.MaxAge = v
		case "expires":
			t, e := time.Parse(time.RFC1123, _val)
			if e != nil {
				return fmt.Errorf("sha: bad cookie value `%s`, expires is not a valid time", item)
			}
			obj.Expires = t
		case "samesite":
			switch strings.ToLower(_val) {
			case "":
				break
			case "lax":
				obj.SameSite = CookieSameSiteLax
			case "strict":
				obj.SameSite = CookieSameSiteStrict
			case "none":
				obj.SameSite = CookieSameSizeNone
			default:
				return fmt.Errorf("sha: bad cookie value `%s`, unexpected samesite value", item)
			}
		default:
			return fmt.Errorf("sha: bad cookie value `%s`, unexpected cookie attribute", item)
		}
	}

	if len(obj.Domain) < 1 {
		obj.Domain = host
	}
	jar.append(obj)
	return nil
}

func (jar *CookieJar) Load(reader io.Reader) error {
	allBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var items []*CookieItem
	if err = jsonx.Unmarshal(allBytes, &items); err != nil {
		return err
	}
	for _, item := range items {
		jar.append(item)
	}
	return nil
}

func (jar *CookieJar) LoadIfExists(fp string) error {
	f, e := os.Open(fp)
	if e != nil {
		if os.IsNotExist(e) {
			return nil
		}
		return e
	}
	defer f.Close()
	return jar.Load(f)
}

func (jar *CookieJar) Save(writer io.Writer) error {
	var items []*CookieItem
	for _, m := range jar.all {
		for _, v := range m {
			if v.Expires.IsZero() && v.MaxAge == 0 {
				continue
			}
			items = append(items, v)
		}
	}
	fmt.Println(items)
	encoder := jsonx.NewEncoder(writer)
	return encoder.Encode(items)
}

func (jar *CookieJar) SaveTo(fp string) error {
	f, e := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0644)
	if e != nil {
		return e
	}
	defer f.Close()
	return jar.Save(f)
}

func (jar *CookieJar) Cookies(domain string) []*CookieItem {
	var lst []*CookieItem

	domain = removePortAndLowercase(domain)
	if net.ParseIP(domain) != nil && !isDomain(domain) {
		for _, item := range jar.all[domain] {
			lst = append(lst, item)
		}
		return lst
	}

	domain = "." + domain
	for d, dm := range jar.all {
		if strings.HasSuffix(domain, d) {
			for _, item := range dm {
				lst = append(lst, item)
			}
		}
	}
	return lst
}

func NewCookieJar() *CookieJar { return &CookieJar{all: map[string]map[string]*CookieItem{}} }
