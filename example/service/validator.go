package service

import (
	"github.com/zzztttkkk/suna/validator"
	"regexp"
)

var spaceRegexp = regexp.MustCompile(`\s+`)
var empty = []byte("")

func init() {
	validator.RegisterFunc(
		"username",
		func(data []byte) ([]byte, bool) {
			v := spaceRegexp.ReplaceAll(data, empty)
			return v, len(v) > 2
		},
		"remove all space, then check the length > 2",
	)

	validator.RegisterRegexp(
		"password",
		regexp.MustCompile(
			"[\\w!@#$%^&*()-=_+\\[\\]{}\\\\|;':\",./<>?]{6,18}",
		),
	)
}
