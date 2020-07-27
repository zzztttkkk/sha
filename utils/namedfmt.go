package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type M map[string]interface{}

var placeholderRegexp = regexp.MustCompile(`{.*?}`)
var nameRegexp = regexp.MustCompile(`^[a-zA-Z_]\w*$`)

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
		fv := "%value"
		if ind != -1 {
			name = strings.TrimSpace(s[:ind])
			fv = "%" + strings.TrimSpace(s[ind+1:])
		}

		if nameRegexp.MatchString(name) {
			nl = append(nl, name)
			prev += fv
		} else {
			prev += s
		}
		bodySs[i] = prev
	}
	return strings.Join(bodySs, ""), nl
}

type _NamedFmtT struct {
	fs string
	ns []string
}

func NewNamedFmt(fs string) *_NamedFmtT {
	f := &_NamedFmtT{}
	f.fs, f.ns = parseFs(fs)
	return f
}

func (f *_NamedFmtT) getVl(v M) (vl []interface{}) {
	for _, n := range f.ns {
		vl = append(vl, v[n])
	}
	return
}

func (f *_NamedFmtT) Render(v M) string {
	return fmt.Sprintf(f.fs, f.getVl(v)...)
}
