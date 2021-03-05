package sha

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"

	"github.com/zzztttkkk/sha/validator"
)

type MuxOption struct {
	Prefix            string `json:"prefix" toml:"prefix"`
	AutoOptions       bool   `json:"auto_options" toml:"auto-options"`
	AutoSlashRedirect bool   `json:"auto_slash_redirect" toml:"auto-slash-redirect"`
}

var defaultMuxOption = MuxOption{
	Prefix:            "",
	AutoOptions:       false,
	AutoSlashRedirect: false,
}

type Mux struct {
	_MiddlewareNode

	prefix            string
	customMethodTrees map[string]*_RouteNode
	stdMethodTrees    []*_RouteNode
	rawMap            map[string]map[string]RequestHandler

	// unlike "Cors", "autoOptions" only runs when the request is same origin and method == "OPTIONS"
	autoOptions       bool
	autoSlashRedirect bool
	docs              map[string]map[string]string

	NotFound RequestHandler
}

var _ Router = (*Mux)(nil)

func NewMux(option *MuxOption, checkOrigin CORSOriginChecker) *Mux {
	if option == nil {
		option = &defaultMuxOption
	}

	mux := &Mux{
		prefix:            option.Prefix,
		customMethodTrees: map[string]*_RouteNode{},
		stdMethodTrees:    make([]*_RouteNode, 10),
		rawMap:            map[string]map[string]RequestHandler{},
		docs:              map[string]map[string]string{},
		autoOptions:       option.AutoOptions,
		autoSlashRedirect: option.AutoSlashRedirect,
	}

	if checkOrigin != nil {
		mux.Use(MiddlewareFunc(func(ctx *RequestCtx, next func()) {
			origin, _ := ctx.Request.Header.Get(HeaderOrigin)
			if len(origin) < 1 {
				next()
				return
			}

			options := checkOrigin(origin)
			if options == nil {
				ctx.SetStatus(StatusForbidden)
				return
			}
			options.writeHeader(ctx, origin)
			next()
		}))
	}
	return mux
}

func (mux *Mux) Print() {
	type _M struct {
		method string
		m      map[string]RequestHandler
	}

	var ms []*_M
	for m, v := range mux.rawMap {
		ms = append(ms, &_M{method: m, m: v})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].method < ms[j].method })

	type _H struct {
		p string
		h RequestHandler
	}

	var buf strings.Builder
	buf.WriteString("Mux:\n")

	for _, v := range ms {
		buf.WriteString(v.method)
		buf.WriteString(":\n")

		var hs []*_H
		for k, hv := range v.m {
			hs = append(hs, &_H{p: k, h: hv})
		}
		sort.Slice(hs, func(i, j int) bool { return hs[i].p < hs[j].p })

		for _, h := range hs {
			buf.WriteString("\tPath: ")
			buf.WriteString(h.p)
			buf.WriteString("\n")
		}
	}

	log.Print(buf.String())
}

func (mux *Mux) doAddHandler(method, path string, handler RequestHandler) {
	mux.doAddHandler1(method, path, handler, false)
}

func (mux *Mux) addToRawMap(method, path string, handler RequestHandler) {
	m := mux.rawMap[method]
	if len(m) < 1 {
		m = map[string]RequestHandler{}
		mux.rawMap[method] = m
	}
	m[path] = handler
}

func (mux *Mux) doAddHandler1(method, path string, handler RequestHandler, doWrap bool) {
	u, e := url.Parse(path)
	if e != nil ||
		len(u.RawQuery) != 0 ||
		len(u.Fragment) != 0 ||
		path[0] != '/' ||
		strings.Contains(path, "//") {
		panic(fmt.Errorf("sha.mux: bad path value `%s`", path))
	}
	defer mux.addToRawMap(method, path, handler)

	if doWrap {
		handler = mux.wrap(handler)
	}

	path = mux.prefix + path
	path = path[1:]

	tree := mux.getOrNewTree(method)
	newNode := tree.addHandler(strings.Split(path, "/"), handler, path)

	if mux.autoSlashRedirect && method != "OPTIONS" {
		autoSlashRedirect(newNode)
	}

	if mux.autoOptions && method != "OPTIONS" {
		mux.doAutoOptions(method, path)
	}

	path = "/" + path
	dh, ok := handler.(Documenter)
	if ok {
		dm := mux.docs[path]
		if dm == nil {
			dm = map[string]string{}
			mux.docs[path] = dm
		}

		dm[method] = dh.Document()
	}
}

func (mux *Mux) doAutoOptions(method, path string) {
	newNode := mux.getOrNewTree("OPTIONS").addHandler(strings.Split(path, "/"), nil, path)
	newNode.addMethod(method)
	if newNode.handler == nil {
		newNode.autoHandler = true
		newNode.handler = RequestHandlerFunc(newNode.handleOptions)
	}
}

func (mux *Mux) AddBranch(prefix string, router Router) {
	v, ok := router.(*_RouteBranch)
	if !ok {
		panic(fmt.Errorf("sha.router: `%v` is not a branch", router))
	}
	v.prefix = prefix
	v.root = mux
	v._MiddlewareNode.parentMwNode = &mux._MiddlewareNode

	v.goDown()
}

func (mux *Mux) getOrNewTree(method string) *_RouteNode {
	ind := 0

	switch method {
	case "GET":
		ind = 1
	case "HEAD":
		ind = 2
	case "POST":
		ind = 3
	case "PUT":
		ind = 4
	case "PATCH":
		ind = 5
	case "DELETE":
		ind = 6
	case "CONNECT":
		ind = 7
	case "OPTIONS":
		ind = 8
	case "TRACE":
		ind = 9
	}

	if ind != 0 {
		if mux.stdMethodTrees[ind] == nil {
			mux.stdMethodTrees[ind] = &_RouteNode{}
		}
		return mux.stdMethodTrees[ind]
	}

	n, ok := mux.customMethodTrees[method]
	if !ok {
		n = &_RouteNode{}
		mux.customMethodTrees[method] = n
	}
	return n
}

func doAutoRectAddSlash(ctx *RequestCtx) {
	ctx.SetStatus(http.StatusMovedPermanently)
	ctx.Request.Path = append(ctx.Request.Path, '/')
	ctx.Response.Header.Set(HeaderLocation, ctx.Request.Path)
}

func doAutoRectDelSlash(ctx *RequestCtx) {
	ctx.SetStatus(http.StatusMovedPermanently)
	path := ctx.Request.Path
	ctx.Response.Header.Set(HeaderLocation, path[:len(path)-1])
}

func autoSlashRedirect(newNode *_RouteNode) {
	var _n *_RouteNode

	if newNode.name == "" {
		_n = newNode.parent
		if _n != nil && _n.handler == nil {
			_n.autoHandler = true
			_n.handler = RequestHandlerFunc(doAutoRectAddSlash)
		}
		return
	}

	_n = newNode.getChild("")
	if _n == nil {
		_n = &_RouteNode{parent: newNode}
		newNode.addChild(_n)
	}

	if _n.handler == nil {
		_n.autoHandler = true
		_n.handler = RequestHandlerFunc(doAutoRectDelSlash)
	}
}

func (mux *Mux) HTTP(method, path string, handler RequestHandler) {
	method = strings.ToUpper(method)
	mux.doAddHandler1(method, path, handler, true)
}

func (mux *Mux) WebSocket(path string, wh WebSocketHandlerFunc) {
	mux.HTTP("get", path, wshToHandler(wh))
}

type _FormRequestHandler struct {
	RequestHandler
	Documenter
}

func (mux *Mux) HTTPWithForm(method, path string, handler RequestHandler, form interface{}) {
	if form == nil {
		mux.HTTP(method, path, handler)
		return
	}

	mux.HTTP(
		method, path,
		&_FormRequestHandler{
			RequestHandler: handler,
			Documenter:     validator.GetRules(reflect.TypeOf(form)),
		},
	)
}

func (mux *Mux) HTTPWithMiddleware(middleware []Middleware, method, path string, handler RequestHandler) {
	mux.HTTP(method, path, handlerWithMiddleware(handler, middleware...))
}

func (mux *Mux) HTTPWithMiddlewareAndForm(middleware []Middleware, method, path string, handler RequestHandler, form interface{}) {
	mux.HTTPWithForm(method, path, handlerWithMiddleware(handler, middleware...), form)
}

var defaultNotFound = RequestHandlerFunc(func(ctx *RequestCtx) { ctx.SetStatus(StatusNotFound) })

func (mux *Mux) Handle(ctx *RequestCtx) {
	defer doRecover(ctx)

	req := &ctx.Request
	notfound := mux.NotFound
	if notfound == nil {
		notfound = defaultNotFound
	}

	var tree *_RouteNode
	if req._method != _MCustom {
		tree = mux.stdMethodTrees[req._method]
	} else {
		tree = mux.customMethodTrees[string(req.Method)]
	}
	if tree == nil {
		notfound.Handle(ctx)
		return
	}

	path := req.Path
	i, n := tree.find(path[1:], &req.Params)
	if n == nil || n.handler == nil {
		notfound.Handle(ctx)
		return
	}

	if len(n.wildcardName) > 0 {
		if i+1 >= len(path) {
			notfound.Handle(ctx)
			return
		}

		path = path[i+1:]
		if path[0] == '/' {
			path = path[1:]
		}
		req.Params.Append(n.wildcardName, path)
	} else if i < len(path)-2 {
		notfound.Handle(ctx)
		return
	}

	n.handler.Handle(ctx)
}

func (mux *Mux) HandleDoc(method, path string, middleware ...Middleware) {
	type Form struct {
		Prefix string `validator:",optional"`
	}

	var handler = func(ctx *RequestCtx) {
		ctx.AutoCompress()

		form := Form{}
		err := ctx.Validate(&form)
		if err != nil {
			panic(err)
		}

		var m = map[string]map[string]string{}
		for k, v := range mux.docs {
			if strings.HasPrefix(k, form.Prefix) {
				m[k] = v
			}
		}

		var paths []string
		for k := range m {
			paths = append(paths, k)
		}
		sort.Strings(paths)

		buf := strings.Builder{}
		for _, path := range paths {
			buf.WriteString(fmt.Sprintf("## path: %s\n\n", path))
			mmap := m[path]
			var methods []string
			for method := range mmap {
				methods = append(methods, method)
			}
			sort.Strings(methods)
			for _, method := range methods {
				buf.WriteString(fmt.Sprintf("### method: %s\n\n", method))
				buf.WriteString(mmap[method])
				buf.WriteString("\n")
			}
		}
		ctx.Response.Header.SetContentType(MIMEMarkdown)
		_, _ = ctx.WriteString(buf.String())
	}

	mux.HTTPWithForm(
		method, path,
		handlerWithMiddleware(RequestHandlerFunc(handler), middleware...),
		Form{},
	)
}

func (mux *Mux) FilePath(fs http.FileSystem, method, path string, autoIndex bool, middleware ...Middleware) {
	mux.HTTP(
		method, path,
		makeFileSystemHandler(fs, path, autoIndex, middleware...),
	)
}

func (mux *Mux) File(fs http.FileSystem, filename, method, path string, middleware ...Middleware) {
	mux.HTTP(method, path, makeFileHandler(fs, filename, middleware...))
}
