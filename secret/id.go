package secret

import (
	"fmt"
	"github.com/zzztttkkk/suna/utils"
	"strconv"
	"strings"
	"time"
)

func DumpId(id int64, seconds int64) string {
	v := fmt.Sprintf("%x@%x:%s", id, time.Now().Unix()+seconds, utils.B2s(RandBytes(12, nil)))
	return fmt.Sprintf("%s|%s", v, utils.B2s(Default.Calc(utils.S2b(v))))
}

func LoadId(s string) (int64, bool) {
	ind := strings.IndexByte(s, '|')
	if ind < 23 {
		return 0, false
	}

	v := s[:ind]
	h := s[ind+1:]

	if len(h) != Default.size {
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

	if !Default.Equal(utils.S2b(v), utils.S2b(h)) {
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
