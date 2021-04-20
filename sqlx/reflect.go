package sqlx

import (
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/zzztttkkk/sha/utils"
	"reflect"
)

type _StructInfo struct {
	groups    map[string]*_StringSet
	immutable _StringSet
}

var reflectMapper = reflectx.NewMapper("db")
var infoCache = map[reflect.Type]*_StructInfo{}

type _StringSet struct {
	data map[string]struct{}
}

func (s *_StringSet) add(v string) {
	if s.data == nil {
		s.data = map[string]struct{}{}
	}
	s.data[v] = struct{}{}
}

func (s *_StringSet) del(v string) {
	delete(s.data, v)
}

func (s *_StringSet) has(v string) bool {
	_, ok := s.data[v]
	return ok
}

func (s *_StringSet) all() []string {
	var lst []string
	for k := range s.data {
		lst = append(lst, k)
	}
	return lst
}

func getStructInfo(t reflect.Type) *_StructInfo {
	ret, ok := infoCache[t]
	if ok {
		return ret
	}

	ret = &_StructInfo{}
	ret.groups = map[string]*_StringSet{}
	ret.groups["*"] = &_StringSet{}

	fmap := reflectMapper.TypeMap(t)
	for _, f := range fmap.Index {
		(ret.groups["*"]).add(f.Name)
		for k, v := range f.Options {
			switch k {
			case "G", "g", "group":
				for _, n := range utils.SplitAndTrim(v, ",") {
					m := ret.groups[n]
					if m == nil {
						m = &_StringSet{}
						ret.groups[n] = m
					}
					m.add(f.Name)
				}
			case "immutable":
				ret.immutable.add(f.Name)
			}
		}
	}
	return ret
}
