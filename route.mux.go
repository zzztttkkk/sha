package suna

import (
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/validator"
	"log"
	"net/http"
	"net/url"
	pathlib "path"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

type _Mux struct {
	_MiddlewareOrg
	_Recover

	prefix string
	m      map[string]*_RouteNode
	trees  []*_RouteNode
	rawMap map[string]map[string]RequestHandler

	optionsMap map[string]*_RouteNode

	// unlike "Cors", "AutoOptions" only runs when the request is same origin and method == "OPTIONS"
	AutoOptions       bool
	AutoSlashRedirect bool
	cors              *CorsOptions
	docs              map[string]map[string]string
}

var _ Router = &_Mux{}

func NewMux(prefix string, corsOptions *CorsOptions) *_Mux {
	mux := &_Mux{
		prefix: prefix,
		m:      map[string]*_RouteNode{},
		rawMap: map[string]map[string]RequestHandler{},
		cors:   corsOptions,
		docs:   map[string]map[string]string{},
	}
	if mux.cors != nil {
		mux.Use(mux.cors)
	}
	return mux
}

func middlewareToString(m Middleware) string {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Func {
		return fmt.Sprintf("F %s", runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name())
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return fmt.Sprintf("S %s.%s", t.PkgPath(), t.Name())
}

var mwType = reflect.TypeOf(_MiddlewareWrapper{})
var fwType = reflect.TypeOf(_FormRequestHandler{})

func handlerToString(h RequestHandler) string {
	t := reflect.TypeOf(h)
	if t.Kind() == reflect.Func {
		return fmt.Sprintf("F %s", runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t == mwType {
		return handlerToString(h.(*_MiddlewareWrapper).raw)
	}

	if t == fwType {
		return handlerToString(h.(*_FormRequestHandler).RequestHandler)
	}

	return fmt.Sprintf("S %s.%s", t.PkgPath(), t.Name())
}

func printRequestHandler(buf *strings.Builder, h RequestHandler, showMiddleware bool) {
	if !showMiddleware {
		buf.WriteString(" Handler: ")
		buf.WriteString(handlerToString(h))
		return
	}

	mw, ok := h.(*_MiddlewareWrapper)
	if ok {
		buf.WriteString("\n\t\tMiddleware:")
		for _, m := range mw.middleware {
			buf.WriteString("\n\t\t\t")
			buf.WriteString(middlewareToString(m))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("\t\tHandler: ")
	buf.WriteString(handlerToString(h))
}

func (mux *_Mux) Print(showHandler, showMiddleware bool) {
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
			if showHandler {
				printRequestHandler(&buf, h.h, showMiddleware)
			}
			buf.WriteString("\n")
		}
	}

	log.Print(buf.String())
}

func (mux *_Mux) doAddHandler(method, path string, handler RequestHandler) {
	mux.doAddHandler1(method, path, handler, false)
}

func (mux *_Mux) addToRawMap(method, path string, handler RequestHandler) {
	m := mux.rawMap[method]
	if len(m) < 1 {
		m = map[string]RequestHandler{}
		mux.rawMap[method] = m
	}
	m[path] = handler
}

func (mux *_Mux) doAddHandler1(method, path string, handler RequestHandler, doWrap bool) {
	u, e := url.Parse(path)
	if e != nil || len(u.RawQuery) != 0 || len(u.Fragment) != 0 {
		panic(fmt.Errorf("suna.mux: bad path value `%s`", path))
	}
	defer mux.addToRawMap(method, path, handler)

	if doWrap {
		handler = mux.wrap(handler)
	}

	if path[0] != '/' {
		panic(fmt.Errorf("suna.router: error path: `%s`", path))
	}
	path = mux.prefix + path
	path = path[1:]

	var newNode *_RouteNode
	tree := mux.getOrNewTree(method)
	tree.addHandler(strings.Split(path, "/"), handler, path, false, "", &newNode)
	tree.freezeParams()

	if mux.AutoSlashRedirect && method != "OPTIONS" {
		mux.autoSlashRedirect(newNode)
	}

	if mux.AutoOptions && method != "OPTIONS" {
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

func (mux *_Mux) doAutoOptions(method, path string) {
	if mux.optionsMap == nil {
		mux.optionsMap = map[string]*_RouteNode{}
	}

	var newNode = mux.optionsMap[path]
	if newNode == nil {
		mux.getOrNewTree("OPTIONS").addHandler(
			strings.Split(path, "/"),
			nil,
			path,
			true,
			method,
			&newNode,
		)
		mux.optionsMap[path] = newNode
	} else {
		newNode.addMethod(method)
	}

	if newNode.handler != nil {
		newNode.handler = RequestHandlerFunc(newNode.handleOptions)
		newNode.autoHandler = true
	}
}

func (mux *_Mux) AddBranch(prefix string, router Router) {
	v, ok := router.(*_RouteBranch)
	if !ok {
		panic(fmt.Errorf("suna.router: `%v` is not a branch", router))
	}
	v.prefix = prefix
	v.root = mux
	v._MiddlewareOrg.parentMOrg = &mux._MiddlewareOrg

	v.sinking()
}

func (mux *_Mux) getOrNewTree(method string) *_RouteNode {
	n, ok := mux.m[method]
	if !ok {
		n = &_RouteNode{}
		mux.m[method] = n
	}
	return n
}

var locationHeader = []byte("Location")

func doAutoRectAddSlash(ctx *RequestCtx) {
	ctx.SetStatus(http.StatusMovedPermanently)
	ctx.Request.Path = append(ctx.Request.Path, '/')
	ctx.Response.Header.Set(locationHeader, ctx.Request.Path)
}

func doAutoRectDelSlash(ctx *RequestCtx) {
	ctx.SetStatus(http.StatusMovedPermanently)
	path := ctx.Request.Path
	ctx.Response.Header.Set(locationHeader, path[:len(path)-1])
}

func (mux *_Mux) autoSlashRedirect(newNode *_RouteNode) {
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

func (mux *_Mux) REST(method, path string, handler RequestHandler) {
	method = strings.ToUpper(method)
	mux.doAddHandler1(method, path, handler, true)
}

func (mux *_Mux) WebSocket(path string, wh WebSocketHandlerFunc) {
	mux.REST("get", path, wshToHandler(wh))
}

type _FormRequestHandler struct {
	RequestHandler
	Documenter
}

func (mux *_Mux) RESTWithForm(method, path string, handler RequestHandler, form interface{}) {
	if form == nil {
		mux.REST(method, path, handler)
		return
	}

	mux.REST(
		method, path,
		&_FormRequestHandler{
			RequestHandler: handler,
			Documenter:     validator.GetRules(reflect.TypeOf(form)),
		},
	)
}

func (mux *_Mux) Handle(ctx *RequestCtx) {
	defer mux.doRecover(ctx)

	req := &ctx.Request
	tree := mux.m[internal.S(req.Method)]
	if tree == nil {
		ctx.SetStatus(http.StatusMethodNotAllowed)
		return
	}

	path := req.Path
	paramsC := 0

	i, n := tree.find(path[1:], &req.Params, &paramsC)
	if n == nil || n.handler == nil {
		ctx.SetStatus(http.StatusNotFound)
		return
	}

	if len(n.params) > 0 && paramsC != len(n.params) {
		ctx.SetStatus(http.StatusNotFound)
		return
	}

	if len(n.wildcardName) > 0 {
		// wildcard path must endswith "/"
		if path[i] == '/' {
			n = n.getChild("")
			if n != nil {
				n.handler.Handle(ctx)
			} else {
				ctx.SetStatus(http.StatusNotFound)
			}
			return
		}
		req.Params.Append(n.wildcardName, path[i+2:])
	} else if i < len(path)-2 {
		ctx.SetStatus(http.StatusNotFound)
		return
	}
	n.handler.Handle(ctx)
}

func (mux *_Mux) HandleDoc(method, path string, middleware ...Middleware) {
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

	mux.RESTWithForm(
		method, path,
		handlerWithMiddleware(RequestHandlerFunc(handler), middleware...),
		Form{},
	)
}

func (mux *_Mux) StaticFile(method, path string, fs http.FileSystem, index bool, middleware ...Middleware) {
	if !strings.HasSuffix(path, "/filename:*") {
		panic(fmt.Errorf("suna.router: bad static path"))
	}

	mux.REST(
		method,
		path,
		handlerWithMiddleware(
			RequestHandlerFunc(
				func(ctx *RequestCtx) {
					filename, _ := ctx.PathParam(internal.B("filename"))
					serveFile(ctx, fs, pathlib.Clean(internal.S(filename)), index)
				},
			),
			middleware...,
		),
	)
}
