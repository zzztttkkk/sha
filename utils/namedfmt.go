package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type M map[string]interface{}

var placeholderRegexp = regexp.MustCompile(`{.*?}`)

//noinspection GoNilness
func parseFs(fs string) (string, []string) {
	var nl []string

	var bodySs []string
	bodySs = append(bodySs, placeholderRegexp.Split(fs, -1)...)

	for i, s := range placeholderRegexp.FindAllString(fs, -1) {
		prev := bodySs[i]
		after := bodySs[i+1]

		if strings.HasSuffix(prev, "{") {
			prev = prev[:len(prev)-1]
			bodySs[i] = prev
		} else {
			s = s[1:]
		}

		if strings.HasPrefix(after, "}") {
			after = after[1:]
			bodySs[i+1] = after
		} else {
			s = s[:len(s)-1]
		}

		ind := strings.Index(s, ":")
		name := s
		fv := "%+v"
		if ind != -1 {
			name = strings.TrimSpace(s[:ind])
			fv = "%" + strings.TrimSpace(s[ind+1:])
		}
		nl = append(nl, name)
		prev += fv
		bodySs[i] = prev
	}
	return strings.Join(bodySs, ""), nl
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
