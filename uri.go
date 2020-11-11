package suna

import (
	"bytes"
)

type URI struct {
	Scheme   []byte
	User     []byte
	Password []byte
	Host     []byte
	Port     int
	Path     []byte
	Query    UrlencodedForm
	Fragment []byte
}

func (uri *URI) init(v []byte, isRequestPath bool) {
	var ind, size int

	if !isRequestPath {
	}

	ind = bytes.IndexByte(v, '?')
	size = len(v)
	if ind < 0 {
		uri.Path = append(uri.Path, v...)
		return
	}
	uri.Path = append(uri.Path, v[:ind]...)
	if ind+1 >= size {
		return
	}
	v = v[ind+1:]
	size = len(v)
	ind = bytes.IndexByte(v, '#')
	var query []byte
	if ind < 0 {
		query = v
	} else {
		query = v[:ind]
		uri.Fragment = v[ind+1:]
	}
	uri.Query.ParseBytes(query)
}
