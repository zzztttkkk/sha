package utils

import (
	"strconv"
	"strings"
)

type Int64Slice []int64

func (p Int64Slice) Len() int { return len(p) }

func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }

func (p Int64Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p Int64Slice) Join(sep string) string {
	buf := strings.Builder{}
	last := len(p) - 1
	for i, v := range p {
		buf.WriteString(strconv.FormatInt(v, 16))
		if i != last {
			buf.WriteString(sep)
		}
	}
	return buf.String()
}
