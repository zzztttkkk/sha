package suna

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/zzztttkkk/router"
	"google.golang.org/grpc"
)

type _LoaderT struct {
	parent   *_LoaderT
	children map[string]*_LoaderT
	name     string
	path     string
	doc      map[string]reflect.Type
	http     func(router router.R)
	grpc     func(server *grpc.Server)
}

func NewLoader() *_LoaderT {
	return &_LoaderT{
		children: make(map[string]*_LoaderT),
		parent:   nil,
	}
}

func (loader *_LoaderT) Path() string {
	if len(loader.path) > 0 {
		return loader.path
	}

	//noinspection GoPreferNilSlice
	s := []string{}
	c := loader

	for c != nil {
		s = append(s, c.name)
		c = c.parent
	}

	l := 0
	r := len(s) - 1
	for l < r {
		s[l], s[r] = s[r], s[l]
		l++
		r--
	}
	loader.path = strings.Join(s, "/")
	return loader.path
}

func (loader *_LoaderT) AddChild(name string, child *_LoaderT) {
	_, exists := loader.children[name]
	if exists {
		panic(fmt.Errorf("suna.loader: `%s`.`%s` is already exists", loader.Path(), name))
	}
	child.parent = loader
	child.name = name
	loader.children[name] = child
}

func (loader *_LoaderT) Get(path string) *_LoaderT {
	cursor := loader
	for _, name := range strings.Split(path, ".") {
		if cursor == nil {
			break
		}
		cursor = cursor.children[name]
	}
	return cursor
}

func (loader *_LoaderT) Http(fn func(router router.R)) {
	if loader.http != nil {
		panic(errors.New("suna.loader: `%s`'s http is registered"))
	}
	loader.http = fn
}

func (loader *_LoaderT) Grpc(fn func(server *grpc.Server)) {
	if loader.grpc != nil {
		panic(errors.New("suna.loader: `%s`'s grpc is registered"))
	}
	loader.grpc = fn
}

func (loader *_LoaderT) Doc(n string, p reflect.Type) {
	loader.doc[n] = p
}

func (loader *_LoaderT) bindHttp(router *router.Router) {
	if loader.http != nil {
		loader.http(router)
	}
	for k, v := range loader.children {
		v.bindHttpGroup(router.Group(k))
	}
}

func (loader *_LoaderT) bindHttpGroup(group *router.Group) {
	if loader.http != nil {
		loader.http(group)
	}
	for k, v := range loader.children {
		v.bindHttpGroup(group.Group(k))
	}
}

func (loader *_LoaderT) bindGrpc(server *grpc.Server) {
	if loader.grpc != nil {
		loader.grpc(server)
	}
	for _, v := range loader.children {
		v.bindGrpc(server)
	}
}
