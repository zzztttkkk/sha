package suna

import (
	"github.com/zzztttkkk/suna/internal"
)

type UrlencodedForm struct {
	internal.Kvs
}

func (form *UrlencodedForm) ParseBytes(p []byte) {
	var key []byte
	var val []byte
	var f bool
	for _, d := range p {
		if d == '&' {
			form.onItem(key, val)
			key = key[:0]
			val = val[:0]
			f = false
			continue
		}
		if d == '=' {
			f = true
			continue
		}
		if f {
			val = append(val, d)
		} else {
			key = append(key, d)
		}
	}
	form.onItem(key, val)
}
