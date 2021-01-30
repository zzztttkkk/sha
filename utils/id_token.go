package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type IDTokenGenerator struct {
	pool *HashPool
}

func NewIDTokenGenerator(pool *HashPool) IDTokenGenerator {
	return IDTokenGenerator{pool: pool}
}

func (g IDTokenGenerator) EncodeID(v int64) string {
	var buf strings.Builder
	buf.WriteString(strconv.FormatInt(v, 16))
	buf.WriteByte(':')
	buf.WriteString(strconv.FormatInt(time.Now().Unix(), 16))

	d := g.pool.Sum(B(buf.String()))
	buf.WriteByte(':')
	buf.Write(d)
	return buf.String()
}

var ErrBadHIDValue = errors.New("sha.utils: bad HID token value")

func (g IDTokenGenerator) DecodeID(v string) (int64, int64, error) {
	ind := strings.LastIndexByte(v, ':')
	if ind < 3 {
		return -1, -1, ErrBadHIDValue
	}

	hv := v[ind+1:]
	if !g.pool.Equal(B(v[:ind]), B(hv)) {
		return -1, -1, ErrBadHIDValue
	}

	v = v[:ind]
	ind = strings.IndexByte(v, ':')
	if ind < 1 {
		return -1, -1, ErrBadHIDValue
	}

	hid, err1 := strconv.ParseInt(v[:ind], 16, 64)
	if err1 != nil {
		return -1, -1, ErrBadHIDValue
	}

	unix, err2 := strconv.ParseInt(v[ind+1:], 16, 64)
	if err2 != nil {
		return -1, -1, ErrBadHIDValue
	}
	return hid, unix, nil
}
