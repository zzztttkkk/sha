package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type MuxOptions struct {
	Prefix                  string                               `json:"prefix" toml:"prefix"`
	DoTrailingSlashRedirect bool                                 `json:"tsr" toml:"tsr"`
	AutoHandleOptions       bool                                 `json:"aho" toml:"aho"`
	NoFound                 func(ctx *RequestCtx)                `json:"-" toml:"-"`
	MethodNotAllowed        func(ctx *RequestCtx)                `json:"-" toml:"-"`
	CORS                    []*CorsOptions                       `json:"cors" toml:"cors"`
	CORSOriginToName        func(origin []byte) string           `json:"-" toml:"-"`
	Recover                 func(ctx *RequestCtx, v interface{}) `json:"recover" toml:"-"`
}

var defaultMuxOption MuxOptions

func init() {
	defaultMuxOption.DoTrailingSlashRedirect = true
	defaultMuxOption.NoFound = func(ctx *RequestCtx) { ctx.Response.statusCode = StatusNotFound }
	defaultMuxOption.MethodNotAllowed = func(ctx *RequestCtx) { ctx.Response.statusCode = StatusMethodNotAllowed }
	defaultMuxOption.Recover = doRecover
	defaultMuxOption.AutoHandleOptions = true
}

type Mux struct {
	_MiddlewareNode

	prefix string

	stdTrees    [10]*_RadixTree
	customTrees map[string]*_RadixTree

	// path -> method -> description
	documents map[string]map[string]validator.Document

	// opt
	doTrailingSlashRedirect bool
	notFound                func(ctx *RequestCtx)
	methodNotAllowed        func(ctx *RequestCtx)
	cors                    map[string]*_CorsOptions
	corsOriginToName        func(origin []byte) string
	recover                 func(ctx *RequestCtx, v interface{})
	autoHandleOptions       bool

	// raw
	// path -> method
	all map[string]map[string]string
}

func (m *Mux) HTTP(method, path string, handler RequestHandler) {
	m.HTTPWithOptions(nil, method, path, handler)
}

var _ Router = (*Mux)(nil)

func (m *Mux) HTTPWithOptions(opt *HandlerOptions, method, path string, handler RequestHandler) {
	var middlewares []Middleware
	var document validator.Document
	if opt != nil {
		middlewares = opt.Middlewares
		document = opt.Document
	}

	method = strings.ToUpper(method)
	ind := 0
	switch method {
	case MethodGet:
		ind = 1
	case MethodHead:
		ind = 2
	case MethodPost:
		ind = 3
	case MethodPut:
		ind = 4
	case MethodPatch:
		ind = 5
	case MethodDelete:
		ind = 6
	case MethodConnect:
		ind = 7
	case MethodOptions:
		ind = 8
	case MethodTrace:
		ind = 9
	}

	var tree *_RadixTree

	if ind != 0 {
		tree = m.stdTrees[ind]
		if tree == nil {
			tree = newRadixTree()
			m.stdTrees[ind] = tree
		}
	} else {
		tree = m.customTrees[method]
		if tree == nil {
			tree = newRadixTree()
			m.customTrees[method] = tree
		}
	}

	if !isAutoOptionsHandler(handler) {
		var ms []Middleware
		ms = append(ms, m._MiddlewareNode.local...)
		ms = append(ms, middlewares...)

		if len(ms) > 0 {
			handler = middlewaresWrap(ms, handler)
		}
	}

	path = m.prefix + path

	tree.Add(path, handler)
	if method != MethodOptions && m.autoHandleOptions {
		m.HTTP(MethodOptions, path, newAutoOptions(method))
	}

	if document != nil {
		m1 := m.documents[path]
		if m1 == nil {
			m1 = map[string]validator.Document{}
			m.documents[path] = m1
		}

		m1[method] = document
	}
}

func (m *Mux) NewGroup(prefix string) Router {
	return &_MuxGroup{
		prefix: prefix,
		mux:    m,
	}
}

func (m *Mux) Websocket(path string, handlerFunc WebSocketHandlerFunc, opt *HandlerOptions) {
	m.HTTPWithOptions(opt, "get", path, wshToHandler(handlerFunc))
}

func makeFileSystemHandler(fs http.FileSystem, autoIndex bool) RequestHandler {
	return RequestHandlerFunc(func(ctx *RequestCtx) {
		fp, _ := ctx.URLParam("filepath")
		serveFileSystem(ctx, fs, filepath.Clean(utils.S(fp)), autoIndex)
	})
}

func (m *Mux) FileSystem(opt *HandlerOptions, method, path string, fs http.FileSystem, autoIndex bool) {
	if !strings.HasSuffix(path, "/{filepath:*}") {
		panic(fmt.Errorf("sha.mux: path must endswith `/{filepath:*}`"))
	}
	m.HTTPWithOptions(
		opt,
		method, path,
		makeFileSystemHandler(fs, autoIndex),
	)
}

func makeFileContentHandler(filepath string) RequestHandler {
	return RequestHandlerFunc(func(ctx *RequestCtx) {
		f, e := os.Open(filepath)
		if e != nil {
			ctx.SetStatus(toHTTPError(e))
			return
		}
		defer f.Close()
		d, err := f.Stat()
		if err != nil {
			ctx.SetStatus(toHTTPError(err))
			return
		}
		serveFileContent(ctx, d.Name(), d.ModTime(), d.Size(), f)
	})
}

func (m *Mux) FileContent(opt *HandlerOptions, method, path, filepath string) {
	if strings.Contains(path, "{") {
		panic(fmt.Errorf("sha.mux: path can not contains `{.*}`"))
	}
	m.HTTPWithOptions(opt, method, path, makeFileContentHandler(filepath))
}

func (m *Mux) Handle(ctx *RequestCtx) {
	defer func() {
		v := recover()
		if v == nil {
			return
		}

		if m.recover != nil {
			m.recover(ctx, v)
			return
		}

		log.Printf("sha.error: %v\n", v)

		ctx.Response.statusCode = StatusInternalServerError
		ctx.Response.ResetBodyBuffer()
		ctx.Response.Header.Reset()
	}()

	var tree *_RadixTree
	if ctx.Request._method != 0 {
		tree = m.stdTrees[ctx.Request._method]
	} else {
		tree = m.customTrees[utils.S(ctx.Request.Method)]
	}

	if tree == nil {
		if m.methodNotAllowed != nil {
			m.methodNotAllowed(ctx)
		} else if m.notFound != nil {
			m.notFound(ctx)
		} else {
			ctx.Response.statusCode = StatusNotFound
		}
		return
	}

	h, tsr := tree.Get(utils.S(ctx.Request.Path), ctx)
	if h == nil {
		if tsr {
			l := len(ctx.Request.Path)
			if ctx.Request.Path[l-1] != '/' {
				item := ctx.Response.Header.Set(HeaderLocation, ctx.Request.Path)
				item.Val = append(item.Val, '/')
			} else {
				ctx.Response.Header.Set(HeaderLocation, ctx.Request.Path[:l-1])
			}
			ctx.Response.statusCode = StatusFound
			return
		}
		if m.notFound != nil {
			m.notFound(ctx)
		} else {
			ctx.Response.statusCode = StatusNotFound
		}
		return
	}

	h.Handle(ctx)
}

func (m *Mux) Print() {

}

func NewMux(opt *MuxOptions) *Mux {
	if opt == nil {
		opt = &defaultMuxOption
	}

	mux := &Mux{
		prefix:      opt.Prefix,
		documents:   map[string]map[string]validator.Document{},
		customTrees: map[string]*_RadixTree{},

		doTrailingSlashRedirect: opt.DoTrailingSlashRedirect,
		methodNotAllowed:        opt.MethodNotAllowed,
		notFound:                opt.NoFound,
		recover:                 opt.Recover,
		autoHandleOptions:       opt.AutoHandleOptions,
	}

	if len(opt.CORS) > 0 {
		if opt.CORSOriginToName == nil {
			panic(fmt.Errorf("sha.mux: nil CORSOriginToName"))
		}
		mux.corsOriginToName = opt.CORSOriginToName
		mux.cors = map[string]*_CorsOptions{}
		for _, co := range opt.CORS {
			mux.cors[co.Name] = newCorsOptions(co)
		}

		mux.Use(MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			origin, _ := ctx.Request.Header.Get(HeaderOrigin)
			if len(origin) < 1 {
				next()
				return
			}

			opt := mux.cors[mux.corsOriginToName(origin)]
			if opt == nil {
				return
			}

			opt.writeHeader(ctx, origin)
			next()
		}))
	}

	return mux
}
