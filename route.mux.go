package suna

import (
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"net/http"
	"strings"
)

type _RouteNode struct {
	params       [][]byte
	paramsStatus int
	raw          string
	wildcardName []byte
	handler      RequestHandler
	parent       *_RouteNode
	name         string
	children     map[string]*_RouteNode
	methods      []byte
	autoHandler  bool
}

func (node *_RouteNode) addMethod(method string) {
	if len(node.methods) > 0 {
		node.methods = append(node.methods, ',', ' ')
	}
	node.methods = append(node.methods, []byte(method)...)
}

var allowHeader = []byte("Allow")

func (node *_RouteNode) handleOptions(ctx *RequestCtx) {
	ctx.Response.Header.Set(allowHeader, node.methods)
}

func (node *_RouteNode) addHandler(
	path []string, handler RequestHandler, raw string,
	isAutoHandler bool, method string,
	newNodePtr **_RouteNode,
) {
	if len(path) < 1 {
		if newNodePtr != nil {
			*newNodePtr = node
		}

		if isAutoHandler {
			node.handler = handler
			if len(method) > 0 {
				node.addMethod(method)
			}
		} else {
			if node.handler != nil && !node.autoHandler {
				panic(fmt.Errorf("suna.router: `%s` conflict with others", raw))
			}
			node.handler = handler
			node.raw = raw
		}
		return
	}

	p := path[0]
	ind := strings.IndexByte(p, ':')
	if ind < 0 {
		if node.paramsStatus == 2 || node.wildcardName != nil {
			panic(fmt.Errorf("suna.router: `%s` conflict with others", raw))
		}

		if node.children == nil {
			node.children = map[string]*_RouteNode{}
		}
		sn := node.children[p]
		if sn == nil {
			sn = &_RouteNode{name: p, parent: node}
			node.children[p] = sn
		}
		sn.addHandler(path[1:], handler, raw, isAutoHandler, method, newNodePtr)
		return
	}

	if ind == 0 {
		if node.paramsStatus == 2 || node.wildcardName != nil {
			panic(fmt.Errorf("suna.router: `%s` conflict with others", raw))
		}

		node.paramsStatus = 1
		node.params = append(node.params, []byte(p[1:]))
		node.addHandler(path[1:], handler, raw, isAutoHandler, method, newNodePtr)
		return
	}

	if len(node.children) != 0 || len(node.params) != 0 || len(node.wildcardName) != 0 {
		panic(fmt.Errorf("suna.router: `%s` conflict with others", raw))
	}

	if !strings.HasSuffix(p, ":*") {
		panic(fmt.Errorf("suna.router: bad path value `%s`", raw))
	}

	node.wildcardName = []byte(p[:len(p)-2])
	node.addHandler(nil, handler, raw, isAutoHandler, method, newNodePtr)
}

func (node *_RouteNode) freezeParams() {
	if node.paramsStatus == 1 {
		node.paramsStatus = 2
	}
	for _, n := range node.children {
		n.freezeParams()
	}
}

func (node *_RouteNode) find(path []byte, kvs *internal.Kvs, paramsC *int) (int, *_RouteNode) {
	n := node
	if len(n.wildcardName) > 0 {
		return 0, n
	}

	var i int
	var b byte
	var key []byte
	params := n.params
	var paramsI int
	var prevI int

	for i, b = range path {
		if b == '/' {
			if len(params) > 0 {
				kvs.Append(params[paramsI], key)
				*paramsC = (*paramsC) + 1
				key = key[:0]
				paramsI++
				if paramsI >= len(params) {
					params = nil
					paramsI = 0
					prevI = i
				}
				continue
			}

			if len(n.children) < 1 {
				return prevI, nil
			}

			n = n.children[internal.S(key)]
			key = key[:0]
			if n == nil {
				return prevI, nil
			}

			if len(n.params) > 0 {
				params = n.params
				prevI = i
				continue
			}

			if len(n.wildcardName) > 0 {
				return i, n
			}

			prevI = i
			continue
		}
		key = append(key, b)
	}
	// path not endswith "/"
	if len(params) > 0 {
		kvs.Append(params[paramsI], key)
		*paramsC = (*paramsC) + 1
		return i, n
	}

	if len(n.children) < 1 {
		return prevI, nil
	}

	n = n.children[internal.S(key)]
	if n == nil {
		return prevI, nil
	}
	return i, n
}

type _Mux struct {
	_MiddlewareOrg
	prefix string
	m      map[string]*_RouteNode
	trees  []*_RouteNode

	optionsMap map[string]*_RouteNode

	// unlike "Cors", "AutoOptions" only runs when the request is same origin and method == "OPTIONS"
	AutoOptions  bool
	AutoRedirect bool
	cors         *CorsOptions
	docs         map[string]map[string]string
}

var _ Router = &_Mux{}

func NewMux(prefix string, corsOptions *CorsOptions) *_Mux {
	mux := &_Mux{
		prefix: prefix,
		m:      map[string]*_RouteNode{},
		trees:  make([]*_RouteNode, 9),
		cors:   corsOptions,
	}
	for i := 0; i < 9; i++ {
		mux.trees[i] = &_RouteNode{}
	}

	if mux.cors != nil {
		if mux.cors.CheckOrigin == nil {
			panic("suna.router: nil `CorsOptions`.`CheckOrigin`")
		}
		mux.Use(mux.cors)
	}
	return mux
}

func (mux *_Mux) doAddHandler(method, path string, handler RequestHandler) {
	mux.doAddHandler1(method, path, handler, false)
}

func (mux *_Mux) doAddHandler1(method, path string, handler RequestHandler, doWrap bool) {
	if doWrap {
		handler = mux.wrap(handler)
	}

	method = strings.ToUpper(method)

	if path[0] != '/' {
		panic(fmt.Errorf("suna.router: error path: `%s`", path))
	}
	path = mux.prefix + path
	path = path[1:]

	var newNode *_RouteNode
	tree := mux.findTree(method)
	tree.addHandler(strings.Split(path, "/"), handler, path, false, "", &newNode)
	tree.freezeParams()

	if mux.AutoRedirect && method != "OPTIONS" {
		mux.autoRedirect(newNode, path)
	}

	if mux.AutoOptions && method != "OPTIONS" {
		mux.doAutoOptions(method, path)
	}

	dh, ok := handler.(DocedRequestHandler)
	if ok {
		dm := mux.docs[method]
		if dm == nil {
			dm = map[string]string{}
			mux.docs[method] = dm
		}
		dm[path] = dh.Document()
	}
}

func (mux *_Mux) doAutoOptions(method, path string) {
	if mux.optionsMap == nil {
		mux.optionsMap = map[string]*_RouteNode{}
	}

	var newNode = mux.optionsMap[path]
	if newNode == nil {
		mux.findTree1("OPTIONS").addHandler(
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

func (mux *_Mux) findTree(method string) *_RouteNode {
	switch method {
	case "GET":
		return mux.trees[0]
	case "HEAD":
		return mux.trees[1]
	case "POST":
		return mux.trees[2]
	case "PUT":
		return mux.trees[3]
	case "DELETE":
		return mux.trees[4]
	case "CONNECT":
		return mux.trees[5]
	case "OPTIONS":
		return mux.trees[6]
	case "TRACE":
		return mux.trees[7]
	case "PATCH":
		return mux.trees[8]
	default:
		n, ok := mux.m[method]
		if !ok {
			n = &_RouteNode{}
			mux.m[method] = n
		}
		return n
	}
}

func (mux *_Mux) findTree1(method string) *_RouteNode {
	switch method {
	case "GET":
		return mux.trees[0]
	case "HEAD":
		return mux.trees[1]
	case "POST":
		return mux.trees[2]
	case "PUT":
		return mux.trees[3]
	case "DELETE":
		return mux.trees[4]
	case "CONNECT":
		return mux.trees[5]
	case "OPTIONS":
		return mux.trees[6]
	case "TRACE":
		return mux.trees[7]
	case "PATCH":
		return mux.trees[8]
	default:
		return mux.m[method]
	}
}

var locationHeader = []byte("Location")

func doAutoRectAddSlash(ctx *RequestCtx) {
	ctx.noBuffer = false
	ctx.WriteError(StdHttpErrors[http.StatusMovedPermanently])
	item := ctx.Response.Header.Set(locationHeader, ctx.Request.Path)
	item.Val = append(item.Val, '/')
}

func doAutoRectDelSlash(ctx *RequestCtx) {
	ctx.noBuffer = false
	ctx.WriteError(StdHttpErrors[http.StatusMovedPermanently])
	path := ctx.Request.Path
	ctx.Response.Header.Set(locationHeader, path[:len(path)-1])
}

func (node *_RouteNode) getChildren(name string) *_RouteNode {
	if len(node.children) == 0 {
		return nil
	}
	return node.children[name]
}

func (mux *_Mux) autoRedirect(newNode *_RouteNode, path string) {
	var _n *_RouteNode

	if newNode.name == "" {
		_n = newNode.parent
		if _n != nil && _n.handler == nil {
			_n.autoHandler = true
			_n.handler = RequestHandlerFunc(doAutoRectAddSlash)
		}
		return
	}

	c := newNode.children
	if len(c) < 1 {
		c = map[string]*_RouteNode{}
		newNode.children = c
	}
	_n = c[""]
	if _n == nil {
		_n = &_RouteNode{parent: newNode}
		c[""] = _n
	}

	if _n.handler == nil {
		_n.autoHandler = true
		_n.handler = RequestHandlerFunc(doAutoRectDelSlash)
	}
}

func (mux *_Mux) AddHandler(method, path string, handler RequestHandler) {
	mux.doAddHandler1(method, path, handler, true)
}

func (mux *_Mux) Handle(ctx *RequestCtx) {
	req := &ctx.Request
	tree := mux.findTree1(internal.S(req.Method))
	if tree == nil {
		ctx.WriteError(StdHttpErrors[http.StatusMethodNotAllowed])
		return
	}

	path := req.Path
	paramsC := 0

	i, n := tree.find(path[1:], &req.Params, &paramsC)
	if n == nil || n.handler == nil {
		ctx.WriteError(StdHttpErrors[http.StatusNotFound])
		return
	}

	if len(n.params) > 0 && paramsC != len(n.params) {
		ctx.WriteError(StdHttpErrors[http.StatusNotFound])
		return
	}

	if len(n.wildcardName) > 0 {
		// wildcard path must endswith "/"
		if path[i] == '/' {
			n = n.getChildren("")
			if n != nil {
				n.handler.Handle(ctx)
			} else {
				ctx.WriteError(StdHttpErrors[http.StatusNotFound])
			}
			return
		}
		req.Params.Append(n.wildcardName, path[i+2:])
	} else if i < len(path)-2 {
		ctx.WriteError(StdHttpErrors[http.StatusNotFound])
		return
	}
	n.handler.Handle(ctx)
}
