package internal

import (
	"strconv"
	"strings"
)

// 0-10: [0, 10]
// 0: [0,0]
// 0-: [0,)
// -10: (,10]
func ParseIntRange(s string) (int64, int64, bool, bool) {
	if len(s) < 1 {
		return 0, 0, false, false
	}

	// -10
	if strings.HasPrefix(s, "-") {
		v, e := strconv.ParseInt(s[1:], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return 0, v, false, true
	}

	// 0-
	if strings.HasSuffix(s, "-") {
		v, e := strconv.ParseInt(s[:len(s)-1], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, 0, true, false
	}

	ss := strings.Split(s, "-")
	if len(ss) == 1 { // 10
		v, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, v, true, true
	}

	if len(ss) != 2 {
		return 0, 0, false, false
	}

	minV, e := strconv.ParseInt(ss[0], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}

	maxV, e := strconv.ParseInt(ss[1], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}
	return minV, maxV, true, true
}

func ParseUintRange(s string) (uint64, uint64, bool, bool) {
	if len(s) < 1 {
		return 0, 0, false, false
	}

	if strings.HasPrefix(s, "-") {
		v, e := strconv.ParseUint(s[1:], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return 0, v, false, true
	}

	if strings.HasSuffix(s, "-") {
		v, e := strconv.ParseUint(s[:len(s)-1], 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, 0, true, false
	}

	ss := strings.Split(s, "-")
	if len(ss) == 1 {
		v, e := strconv.ParseUint(s, 10, 64)
		if e != nil {
			return 0, 0, false, false
		}
		return v, v, true, true
	}
	if len(ss) != 2 {
		return 0, 0, false, false
	}

	minV, e := strconv.ParseUint(ss[0], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}

	maxV, e := strconv.ParseUint(ss[1], 10, 64)
	if e != nil {
		return 0, 0, false, false
	}
	return minV, maxV, true, true
}

func ParseFloatRange(s string) (float64, float64, bool, bool) {
	if len(s) < 1 {
		return 0, 0, false, false
	}

	if strings.HasPrefix(s, "-") {
		v, e := strconv.ParseFloat(s[1:], 10)
		if e != nil {
			return 0, 0, false, false
		}
		return 0, v, false, true
	}

	if strings.HasSuffix(s, "-") {
		v, e := strconv.ParseFloat(s[:len(s)-1], 10)
		if e != nil {
			return 0, 0, false, false
		}
		return v, 0, true, false
	}

	ss := strings.Split(s, "-")
	if len(ss) == 1 {
		v, e := strconv.ParseFloat(s, 10)
		if e != nil {
			return 0, 0, false, false
		}
		return v, v, true, true
	}
	if len(ss) != 2 {
		return 0, 0, false, false
	}

	minV, e := strconv.ParseFloat(ss[0], 10)
	if e != nil {
		return 0, 0, false, false
	}

	maxV, e := strconv.ParseFloat(ss[1], 10)
	if e != nil {
		return 0, 0, false, false
	}
	return minV, maxV, true, true
}

func StrSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
