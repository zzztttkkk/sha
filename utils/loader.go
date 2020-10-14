package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/router"
)

type Loader struct {
	parent   *Loader
	children map[string]*Loader
	name     string
	path     string
	httpFns  []func(router router.Router)
}

func NewLoader() *Loader {
	return &Loader{
		children: make(map[string]*Loader),
		parent:   nil,
	}
}

func (loader *Loader) Path() string {
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
	loader.path = strings.Join(s, "")
	return loader.path
}

func (loader *Loader) AddChild(name string, child *Loader) {
	_, exists := loader.children[name]
	if exists {
		panic(fmt.Errorf("suna.utils.loader: `%s`.`%s` is already exists", loader.Path(), name))
	}
	child.parent = loader
	child.name = name
	loader.children[name] = child
}

func (loader *Loader) Get(path string) *Loader {
	cursor := loader
	for _, name := range strings.Split(path, ".") {
		if cursor == nil {
			break
		}
		cursor = cursor.children[name]
	}
	return cursor
}

func (loader *Loader) Http(fn func(router router.Router)) {
	loader.httpFns = append(loader.httpFns, fn)
}

func (loader *Loader) bindHttp(router router.Router) {
	for _, fn := range loader.httpFns {
		fn(router)
	}
	for k, v := range loader.children {
		v.bindHttp(router.SubGroup(k))
	}
}

func (loader *Loader) RunAsHttpServer(root *router.Root, addr, certFile, keyFile string) {
	loader.bindHttp(root)

	glog := AcquireGroupLogger("Router")
	for method, paths := range root.List() {
		for _, path := range paths {
			glog.Println(fmt.Sprintf("%s: %s", method, path))
		}
	}
	glog.Free()

	server := &fasthttp.Server{
		Handler:               root.Handler,
		NoDefaultServerHeader: true,
		NoDefaultDate:         true,
	}

	var err error
	if len(certFile) > 0 && len(keyFile) > 0 {
		err = server.ListenAndServeTLS(addr, certFile, keyFile)
	} else {
		err = server.ListenAndServe(addr)
	}
	log.Fatal(err)
}
