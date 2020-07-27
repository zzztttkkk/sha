package utils

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"reflect"
	"strings"

	"github.com/zzztttkkk/router"
	"google.golang.org/grpc"
)

type Loader struct {
	parent   *Loader
	children map[string]*Loader
	name     string
	path     string
	doc      map[string]reflect.Type
	http     func(router router.Router)
	grpc     func(server *grpc.Server)
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
	loader.path = strings.Join(s, "/")
	return loader.path
}

func (loader *Loader) AddChild(name string, child *Loader) {
	_, exists := loader.children[name]
	if exists {
		panic(fmt.Errorf("suna.loader: `%s`.`%s` is already exists", loader.Path(), name))
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
	if loader.http != nil {
		panic(errors.New("suna.loader: `%s`'s http is registered"))
	}
	loader.http = fn
}

func (loader *Loader) Grpc(fn func(server *grpc.Server)) {
	if loader.grpc != nil {
		panic(errors.New("suna.loader: `%s`'s grpc is registered"))
	}
	loader.grpc = fn
}

func (loader *Loader) Doc(n string, p reflect.Type) {
	loader.doc[n] = p
}

func (loader *Loader) bindHttp(router router.Router) {
	if loader.http != nil {
		loader.http(router)
	}
	for k, v := range loader.children {
		v.bindHttp(router.SubGroup(k))
	}
}

func (loader *Loader) bindGrpc(server *grpc.Server) {
	if loader.grpc != nil {
		loader.grpc(server)
	}
	for _, v := range loader.children {
		v.bindGrpc(server)
	}
}

func (loader *Loader) RunAsHttpServer(root router.Router, address string) {
	r, ok := root.(*router.Root)
	if !ok {
		panic("")
	}

	loader.bindHttp(root)

	glog := AcquireGroupLogger("Root")
	for method, paths := range r.List() {
		for _, path := range paths {
			glog.Println(fmt.Sprintf("%s: %s", method, path))
		}
	}
	ReleaseGroupLogger(glog)
	log.Fatal(fasthttp.ListenAndServe(address, r.Handler))
}

func (loader *Loader) RunAsGrpcServer(server *grpc.Server, address string) {
	loader.bindGrpc(server)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	log.Fatal(server.Serve(listener))
}
