package secret

import (
	"encoding/base64"
	"fmt"
	"github.com/savsgio/gotils"
	"strconv"
	"strings"
	"time"
)

func DumpId(id int64, seconds int64) string {
	v := fmt.Sprintf("%x@%x:%s", id, time.Now().Unix()+seconds, gotils.B2S(RandBytes(8, nil)))
	return base64.StdEncoding.EncodeToString(gotils.S2B(fmt.Sprintf("%s|%s", v, gotils.B2S(_Default.Calc(gotils.S2B(v))))))
}

func LoadId(s string) (int64, bool) {
	_v, e := base64.StdEncoding.DecodeString(s)
	if e != nil {
		return 0, false
	}

	s = gotils.B2S(_v)

	ind := strings.IndexByte(gotils.B2S(_v), '|')
	if ind < 19 {
		return 0, false
	}

	v := s[:ind]
	h := s[ind+1:]

	if len(h) != _Default.size {
		return 0, false
	}

	ind = strings.IndexByte(v, ':')
	if ind < 10 {
		return 0, false
	}

	nt := v[:ind]
	ind = strings.IndexByte(nt, '@')
	if ind < 1 {
		return 0, false
	}

	if !_Default.Equal(gotils.S2B(v), gotils.S2B(h)) {
		return 0, false
	}

	i, e := strconv.ParseInt(nt[:ind], 16, 64)
	if e != nil {
		return 0, false
	}

	t, _ := strconv.ParseInt(nt[ind+1:], 16, 64)
	if time.Now().Unix() > t {
		return 0, false
	}
	return i, true
}
