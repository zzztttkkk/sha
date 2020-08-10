package utils

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"google.golang.org/grpc"
	"log"
	"net"
	"reflect"
	"strings"
)

type Loader struct {
	parent   *Loader
	children map[string]*Loader
	name     string
	path     string
	doc      map[string]reflect.Type
	httpFns  []func(router router.Router)
	grpcFns  []func(server *grpc.Server)
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
	loader.httpFns = append(loader.httpFns, fn)
}

func (loader *Loader) Grpc(fn func(server *grpc.Server)) {
	loader.grpcFns = append(loader.grpcFns, fn)
}

func (loader *Loader) Doc(n string, p reflect.Type) {
	loader.doc[n] = p
}

func (loader *Loader) bindHttp(router router.Router) {
	for _, fn := range loader.httpFns {
		fn(router)
	}
	for k, v := range loader.children {
		v.bindHttp(router.SubGroup(k))
	}
}

func (loader *Loader) bindGrpc(server *grpc.Server) {
	for _, fn := range loader.grpcFns {
		fn(server)
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

	server := &fasthttp.Server{
		Handler:               r.Handler,
		NoDefaultServerHeader: true,
		NoDefaultDate:         true,
	}
	log.Fatal(server.ListenAndServe(address))
}

func (loader *Loader) RunAsGrpcServer(server *grpc.Server, address string) {
	loader.bindGrpc(server)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	log.Fatal(server.Serve(listener))
}
