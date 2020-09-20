package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/savsgio/gotils"
)

type M map[string]interface{}

var placeholderRegexp = regexp.MustCompile("\\${.*?}")

func parseFs(fs string) (string, []string) {
	var n []string
	var f []string

	prev := 0
	for _, v := range placeholderRegexp.FindAllSubmatchIndex(gotils.S2B(fs), -1) {
		begin, end := v[0], v[1]
		f = append(f, fs[prev:begin])
		prev = end
		txt := fs[begin+2 : end-1]
		ind := strings.IndexByte(txt, ':')
		if ind < 0 {
			f = append(f, "%v")
			n = append(n, txt)
			continue
		}
		f = append(f, "%"+txt[ind+1:])
		n = append(n, txt[:ind])
	}
	return strings.Join(append(f, fs[prev:]), ""), n
}

type NamedFmt struct {
	fs    string
	Names []string
}

func NewNamedFmt(fs string) *NamedFmt {
	f := &NamedFmt{}
	f.fs, f.Names = parseFs(fs)
	return f
}

func (f *NamedFmt) getVl(v M) (vl []interface{}) {
	for _, n := range f.Names {
		vl = append(vl, v[n])
	}
	return
}

func (f *NamedFmt) Render(v M) string {
	return fmt.Sprintf(f.fs, f.getVl(v)...)
}
