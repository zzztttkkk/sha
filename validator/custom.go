package validator

import "regexp"

var regexpMap = map[string]*regexp.Regexp{}
var bytesFilterMap = map[string]func(v []byte) ([]byte, bool){}
var bytesFilterDescriptionMap = map[string]string{}

func RegisterRegexp(name string, rp *regexp.Regexp) {
	regexpMap[name] = rp
}

func RegisterBytesFilter(name string, fn func(v []byte) ([]byte, bool)) {
	bytesFilterMap[name] = fn
}

func RegisterBytesFilterWithDescription(name string, fn func(v []byte) ([]byte, bool), description string) {
	bytesFilterMap[name] = fn
	bytesFilterDescriptionMap[name] = description
}
