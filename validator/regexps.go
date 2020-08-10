package validator

import "regexp"

var regexpMap = map[string]*regexp.Regexp{}

func RegisterRegexp(name string, reg *regexp.Regexp) {
	regexpMap[name] = reg
}
