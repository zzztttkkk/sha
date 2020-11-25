package sqlx

import (
	"github.com/jmoiron/sqlx/reflectx"
	"reflect"
	"strings"
)

type _StructInfo struct {
	groups map[string]map[string]struct{}
}

var reflectMapper = reflectx.NewMapper("db")
var infoCache = map[reflect.Type]*_StructInfo{}

func getStructInfo(t reflect.Type) *_StructInfo {
	ret, ok := infoCache[t]
	if ok {
		return ret
	}

	ret = &_StructInfo{}
	ret.groups = map[string]map[string]struct{}{}
	ret.groups["*"] = map[string]struct{}{}

	fmap := reflectMapper.TypeMap(t)
	for _, f := range fmap.Index {
		ret.groups["*"][f.Name] = struct{}{}
		for k, v := range f.Options {
			switch k {
			case "G", "g", "group":
				for _, n := range strings.Split(v, "|") {
					m := ret.groups[n]
					if m == nil {
						m = map[string]struct{}{}
						ret.groups[n] = m
					}
					m[f.Name] = struct{}{}
				}
			}
		}
	}
	return ret
}
