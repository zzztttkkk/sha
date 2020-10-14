package validator

import "regexp"

var regexpMap = map[string]*regexp.Regexp{}

// RegisterRegexp you must call this function before calling other functions
func RegisterRegexp(name string, reg *regexp.Regexp) {
	regexpMap[name] = reg
}
