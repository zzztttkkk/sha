package sha

import (
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
	"regexp"
)

type _MuxItem struct {
	opt     *HandlerOptions
	path    string
	method  string
	handler RequestHandler
}

type MuxGroup struct {
	_MiddlewareNode

	prefix string
	parent *MuxGroup
	mux    *Mux

	lazy  bool
	cache []*_MuxItem
}

func (m *MuxGroup) Websocket(path string, handlerFunc WebsocketHandlerFunc, opt *HandlerOptions) {
	m.HTTPWithOptions(opt, "get", path, wshToHandler(handlerFunc))
}

func (m *MuxGroup) FileSystem(opt *HandlerOptions, method, path string, fs http.FileSystem, autoIndex bool) {
	m.HTTPWithOptions(opt, method, path, makeFileSystemHandler(path, fs, autoIndex))
}

func (m *MuxGroup) File(opt *HandlerOptions, method, path, filepath string) {
	m.HTTPWithOptions(opt, method, path, makeFileContentHandler(path, filepath))
}

func (m *MuxGroup) HTTP(method, path string, handler RequestHandler) {
	m.HTTPWithOptions(nil, method, path, handler)
}

var _ Router = (*MuxGroup)(nil)

func (m *MuxGroup) HTTPWithOptions(opt *HandlerOptions, method, path string, handler RequestHandler) {
	if m.lazy {
		opt = m.copyMiddleware(opt)
		m.cache = append(m.cache, &_MuxItem{opt: opt, path: m.prefix + path, method: method, handler: handler})
		return
	}

	m.add(method, m.prefix+path, handler, opt)
}

func (m *MuxGroup) HTTPWithForm(method, path string, handler RequestHandler, form interface{}) {
	m.HTTPWithOptions(&HandlerOptions{Document: validator.NewDocument(form, validator.Undefined)}, method, path, handler)
}

func (m *MuxGroup) copyMiddleware(opt *HandlerOptions) *HandlerOptions {
	if opt == nil {
		opt = &HandlerOptions{}
	}
	om := opt.Middlewares
	opt.Middlewares = make([]Middleware, 0)
	opt.Middlewares = append(opt.Middlewares, m._MiddlewareNode.local...)
	opt.Middlewares = append(opt.Middlewares, om...)
	return opt
}

func (m *MuxGroup) add(method, path string, handler RequestHandler, opt *HandlerOptions) {
	opt = m.copyMiddleware(opt)
	if m.parent != nil {
		m.parent.add(method, path, handler, opt)
		return
	}
	m.mux.HTTPWithOptions(opt, method, path, handler)
}

var prefixReg = regexp.MustCompile("/[a-zA-Z_]*")

func checkPrefix(v string) string {
	if len(v) == 0 {
		return v
	}
	if !prefixReg.MatchString(v) {
		panic(fmt.Errorf("sha.router: bad prefix `%s`", v))
	}
	return v
}

func (m *MuxGroup) NewGroup(prefix string) Router {
	return &MuxGroup{
		prefix: checkPrefix(prefix),
		parent: m,
		mux:    m.mux,
	}
}

func NewRouteGroup(prefix string) *MuxGroup { return &MuxGroup{prefix: checkPrefix(prefix), lazy: true} }

func (m *MuxGroup) BindTo(router Router) {
	if !m.lazy {
		panic(errors.New("sha.mux: bad group"))
	}

	for _, item := range m.cache {
		router.HTTPWithOptions(item.opt, item.method, item.path, item.handler)
	}
	m.cache = nil
}
