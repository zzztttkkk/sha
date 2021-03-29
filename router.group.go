package sha

import (
	"fmt"
	"net/http"
	"regexp"
)

type _MuxGroup struct {
	_MiddlewareNode

	prefix string
	parent *_MuxGroup
	mux    *Mux
}

func (m *_MuxGroup) Websocket(path string, handlerFunc WebsocketHandlerFunc, opt *HandlerOptions) {
	m.HTTPWithOptions(opt, "get", path, wshToHandler(handlerFunc))
}

func (m *_MuxGroup) FileSystem(opt *HandlerOptions, method, path string, fs http.FileSystem, autoIndex bool) {
	m.HTTPWithOptions(opt, method, path, makeFileSystemHandler(path, fs, autoIndex))
}

func (m *_MuxGroup) File(opt *HandlerOptions, method, path, filepath string) {
	m.HTTPWithOptions(opt, method, path, makeFileContentHandler(path, filepath))
}

func (m *_MuxGroup) HTTP(method, path string, handler RequestHandler) {
	m.HTTPWithOptions(nil, method, path, handler)
}

var _ Router = (*_MuxGroup)(nil)

func (m *_MuxGroup) HTTPWithOptions(opt *HandlerOptions, method, path string, handler RequestHandler) {
	m.add(nil, method, m.prefix+path, handler, opt)
}

func (m *_MuxGroup) add(childMiddlewares []Middleware, method, path string, handler RequestHandler, opt *HandlerOptions) {
	var ms []Middleware
	ms = append(ms, m.local...)
	ms = append(ms, childMiddlewares...)

	if m.parent != nil {
		m.parent.add(ms, method, path, handler, opt)
		return
	}

	if len(ms) != 0 {
		if opt == nil {
			opt = &HandlerOptions{
				Middlewares: ms,
			}
		} else {
			nopt := &HandlerOptions{}
			nopt.Document = opt.Document
			nopt.Middlewares = append(nopt.Middlewares, ms...)
			nopt.Middlewares = append(nopt.Middlewares, opt.Middlewares...)
			opt = nopt
		}
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

func (m *_MuxGroup) NewGroup(prefix string) Router {
	return &_MuxGroup{
		prefix: checkPrefix(prefix),
		parent: m,
		mux:    m.mux,
	}
}
