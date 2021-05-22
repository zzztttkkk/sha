package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MuxOptions struct {
	Prefix                  string                               `json:"prefix" toml:"prefix"`
	DoTrailingSlashRedirect bool                                 `json:"tsr" toml:"tsr"`
	AutoHandleOptions       bool                                 `json:"aho" toml:"aho"`
	NoFound                 func(ctx *RequestCtx)                `json:"-" toml:"-"`
	MethodNotAllowed        func(ctx *RequestCtx)                `json:"-" toml:"-"`
	CORS                    []*CorsOptions                       `json:"cors" toml:"cors"`
	CORSSKip                bool                                 `json:"cors_skip" toml:"cors-skip"`
	CORSOriginToName        func(origin []byte) string           `json:"-" toml:"-"`
	Recover                 func(ctx *RequestCtx, v interface{}) `json:"recover" toml:"-"`
}

var defaultMuxOption MuxOptions

func init() {
	defaultMuxOption.DoTrailingSlashRedirect = true
	defaultMuxOption.NoFound = func(ctx *RequestCtx) { ctx.Response.SetStatusCode(StatusNotFound) }
	defaultMuxOption.Recover = doRecover
	defaultMuxOption.AutoHandleOptions = true
}

type Mux struct {
	_MiddlewareNode
	Opts MuxOptions

	stdTrees    [10]*_RadixTree
	customTrees map[string]*_RadixTree

	// path -> method -> description
	documents         map[string]map[string]validator.Document
	documentsTagIndex map[string]map[validator.Document]struct{}

	cors map[string]*_CorsOptions

	// raw
	// path -> method
	all map[string]map[string]string
}

func (m *Mux) Frozen() { m.frozen = true }

func (m *Mux) AddGroup(group *MuxGroup) { group.BindTo(m) }

func (m *Mux) HTTP(method, path string, handler RequestHandler) {
	m.HTTPWithOptions(nil, method, path, handler)
}

var _ Router = (*Mux)(nil)

func isFileSystemHandler(h RequestHandler) bool {
	_, ok := h.(*_FileSystemHandler)
	return ok
}

func isFileContentHandler(h RequestHandler) bool {
	_, ok := h.(*_FileContentHandler)
	return ok
}

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

	rawHandler := handler

	if !isAutoOptionsHandler(handler) {
		var ms []Middleware
		ms = append(ms, m._MiddlewareNode.local...)
		ms = append(ms, middlewares...)

		if len(ms) > 0 {
			handler = middlewaresWrap(ms, handler)
		}
	}

	opts := &m.Opts
	path = opts.Prefix + path

	tree.Add(path, handler)
	if method != MethodOptions && opts.AutoHandleOptions {
		m.HTTP(MethodOptions, path, newAutoOptions(method))
	}

	if document != nil {
		mapAppend(m.documents, path, method, document)

		for _, tag := range document.Tags() {
			m2 := m.documentsTagIndex[tag]
			if m2 == nil {
				m2 = map[validator.Document]struct{}{}
				m.documentsTagIndex[tag] = m2
			}
			m2[document] = struct{}{}
		}
	}

	if isAutoOptionsHandler(rawHandler) {
		return
	}

	m2 := m.all[path]
	if m2 == nil {
		m2 = map[string]string{}
		m.all[path] = m2
	}

	handlerDesc := ""
	if isFileSystemHandler(rawHandler) {
		handlerDesc = fmt.Sprintf("FileSystem %s, auto_index=%v", rawHandler.(*_FileSystemHandler).fs, rawHandler.(*_FileSystemHandler).autoIndex)
	} else if isFileContentHandler(rawHandler) {
		handlerDesc = fmt.Sprintf("File %s", rawHandler.(*_FileContentHandler).fp)
	}
	m2[method] = handlerDesc
}

func (m *Mux) HTTPWithForm(method, path string, handler RequestHandler, form interface{}) {
	m.HTTPWithOptions(&HandlerOptions{Document: validator.NewDocument(form, validator.Undefined)}, method, path, handler)
}

func (m *Mux) NewGroup(prefix string) Router {
	return &MuxGroup{
		prefix: checkPrefix(prefix),
		mux:    m,
	}
}

func (m *Mux) Websocket(path string, handlerFunc WebsocketHandlerFunc, opt *HandlerOptions) {
	m.HTTPWithOptions(opt, "get", path, wshToHandler(handlerFunc))
}

type _FileSystemHandler struct {
	fs        http.FileSystem
	autoIndex bool
}

func (fh *_FileSystemHandler) Handle(ctx *RequestCtx) {
	fp, _ := ctx.Request.URL.Params.Get("filepath")
	serveFileSystem(ctx, fh.fs, filepath.Clean(utils.S(fp)), fh.autoIndex)
}

func makeFileSystemHandler(path string, fs http.FileSystem, autoIndex bool) RequestHandler {
	if !strings.HasSuffix(path, "/{filepath:*}") {
		panic(fmt.Errorf("sha.mux: path must endswith `/{filepath:*}`"))
	}
	return &_FileSystemHandler{autoIndex: autoIndex, fs: fs}
}

func (m *Mux) FileSystem(opt *HandlerOptions, method, path string, fs http.FileSystem, autoIndex bool) {
	m.HTTPWithOptions(
		opt,
		method, path,
		makeFileSystemHandler(path, fs, autoIndex),
	)
}

type _FileContentHandler struct {
	fp string
}

func (fh *_FileContentHandler) Handle(ctx *RequestCtx) {
	f, e := os.Open(fh.fp)
	if e != nil {
		ctx.Response.SetStatusCode(_FSErrToHTTPError(e))
		return
	}
	defer f.Close()
	d, err := f.Stat()
	if err != nil {
		ctx.Response.SetStatusCode(_FSErrToHTTPError(err))
		return
	}
	serveFileContent(ctx, d.Name(), d.ModTime(), d.Size(), f)
}

func makeFileContentHandler(path, filepath string) RequestHandler {
	if strings.Contains(path, "{") {
		panic(fmt.Errorf("sha.mux: path can not contains `{.*}`"))
	}
	return &_FileContentHandler{fp: filepath}
}

func (m *Mux) File(opt *HandlerOptions, method, path, filepath string) {
	m.HTTPWithOptions(opt, method, path, makeFileContentHandler(path, filepath))
}

func (m *Mux) getTree(ctx *RequestCtx) *_RadixTree {
	var tree *_RadixTree
	if ctx.Request._method > 1 {
		tree = m.stdTrees[ctx.Request._method-1]
	} else {
		tree = m.customTrees[utils.S(ctx.Request.Method())]
	}
	return tree
}

func (m *Mux) onNotFound(ctx *RequestCtx) {
	opts := &m.Opts

	if opts.MethodNotAllowed != nil && ctx.Request._method != _MOptions {
		optionsTree := m.stdTrees[_MOptions-1]
		if optionsTree != nil {
			h, _ := optionsTree.Get(ctx.Request.Path(), ctx)
			if h != nil {
				ctx.Response.SetStatusCode(StatusMethodNotAllowed)
				h.Handle(ctx)
				return
			}
		} else {
			opts.MethodNotAllowed(ctx)
			return
		}
	}

	if opts.NoFound != nil {
		opts.NoFound(ctx)
		return
	}
	ctx.Response.statusCode = StatusNotFound
}

func (m *Mux) Handle(ctx *RequestCtx) {
	defer func() {
		v := recover()
		if v == nil {
			v = ctx.err
			if v == nil {
				return
			}
		}

		if m.Opts.Recover != nil {
			m.Opts.Recover(ctx, v)
			return
		}

		log.Printf("sha.error: %v\n", v)

		ctx.Response.SetStatusCode(StatusInternalServerError)
		ctx.Response.ResetBody()
		ctx.Response.Header().Reset()
	}()

	req := &ctx.Request
	res := &ctx.Response

	tree := m.getTree(ctx)
	if tree == nil {
		m.onNotFound(ctx)
		return
	}

	path := req.Path()

	h, tsr := tree.Get(path, ctx)
	if h == nil {
		if tsr {
			l := len(path)
			if path[l-1] != '/' {
				item := res.Header().SetString(HeaderLocation, path)
				item.Val = append(item.Val, '/')
			} else {
				res.Header().SetString(HeaderLocation, path[:l-1])
			}
			res.SetStatusCode(StatusFound)
			return
		}
		m.onNotFound(ctx)
		return
	}
	h.Handle(ctx)
}

func (m *Mux) String() string {
	var buf strings.Builder
	var ps []string
	for p := range m.all {
		ps = append(ps, p)
	}
	sort.Slice(ps, func(i, j int) bool { return strings.ToUpper(ps[i]) < strings.ToUpper(ps[j]) })

	for _, p := range ps {
		pm := m.all[p]

		buf.WriteString(fmt.Sprintf("Path: %s\n", p))

		var ms []string
		for me := range pm {
			ms = append(ms, me)
		}
		sort.Slice(ms, func(i, j int) bool { return ms[i] < ms[j] })

		if len(ms) > 0 {
			buf.WriteByte('\t')
		}

		for i, me := range ms {
			h := pm[me]
			buf.WriteString(me)
			if len(h) > 0 {
				buf.WriteString("(")
				buf.WriteString(h)
				buf.WriteString(")")
			}
			if i < len(ms)-1 {
				buf.WriteString(", ")
			}
		}

		buf.WriteByte('\n')
	}

	return buf.String()
}

func NewMux(opts *MuxOptions) *Mux {
	if opts == nil {
		opts = &defaultMuxOption
	}

	mux := &Mux{
		Opts:        *opts,
		documents:   map[string]map[string]validator.Document{},
		customTrees: map[string]*_RadixTree{},
		all:         map[string]map[string]string{},
	}

	opts = &mux.Opts
	if len(opts.CORS) > 0 {
		if opts.CORSOriginToName == nil {
			panic(fmt.Errorf("sha.mux: nil CORSOriginToName"))
		}
		mux.cors = map[string]*_CorsOptions{}
		for _, co := range opts.CORS {
			mux.cors[co.Name] = newCorsOptions(co)
		}

		mux.Use(MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			origin, _ := ctx.Request.Header().Get(HeaderOrigin)
			if len(origin) < 1 {
				next()
				return
			}

			if opts.CORSSKip {
				allowAllCorsOption.writeHeader(ctx, origin)
				next()
				return
			}

			corsOptions := mux.cors[opts.CORSOriginToName(origin)]
			if corsOptions == nil {
				return
			}

			corsOptions.writeHeader(ctx, origin)
			next()
		}))
	}

	return mux
}
