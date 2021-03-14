package sha

import "github.com/zzztttkkk/sha/validator"

type _Docs struct {
	version string
	m       map[string]map[string]validator.Document
}

func (d *_Docs) add(path, method string, doc validator.Document) {
	if d.m == nil {
		d.m = map[string]map[string]validator.Document{}
	}

	_m, ok := d.m[path]
	if !ok {
		_m = map[string]validator.Document{}
		d.m[path] = _m
	}
	_m[method] = doc
}
