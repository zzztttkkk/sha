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
	children     map[string]*_RouteNode
}

func (node *_RouteNode) addHandler(path []string, handler RequestHandler, raw string) {
	if len(path) < 1 {
		node.handler = handler
		node.raw = raw
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
			sn = &_RouteNode{}
			node.children[p] = sn
		}
		sn.addHandler(path[1:], handler, raw)
		return
	}

	if ind == 0 {
		if node.paramsStatus == 2 || node.wildcardName != nil {
			panic(fmt.Errorf("suna.router: `%s` conflict with others", raw))
		}

		node.paramsStatus = 1
		node.params = append(node.params, []byte(p[1:]))
		node.addHandler(path[1:], handler, raw)
		return
	}

	if len(node.children) != 0 || len(node.params) != 0 || len(node.wildcardName) != 0 {
		panic(fmt.Errorf("suna.router: `%s` conflict with others", raw))
	}

	if !strings.HasSuffix(p, ":*") {
		panic(fmt.Errorf("suna.router: bad path value `%s`", raw))
	}

	node.wildcardName = []byte(p[:len(p)-2])
	node.handler = handler
	node.raw = raw
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

type Router interface {
	AddHandler(method, path string, handler RequestHandler)
	Branch(prefix string) Router
}

type _Mux struct {
	prefix string
	m      map[string]*_RouteNode
	trees  []*_RouteNode
}

func newMux(prefix string) *_Mux {
	mux := &_Mux{
		prefix: prefix,
		m:      map[string]*_RouteNode{},
		trees:  make([]*_RouteNode, 9),
	}
	for i := 0; i < 9; i++ {
		mux.trees[i] = &_RouteNode{}
	}
	return mux
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

func (mux *_Mux) AddHandler(method, path string, handler RequestHandler) {
	if path[0] != '/' {
		panic(fmt.Errorf("suna.router: error path: `%s`", path))
	}
	path = mux.prefix + path
	path = path[1:]

	tree := mux.findTree(method)
	tree.addHandler(strings.Split(path, "/"), handler, path)
	tree.freezeParams()
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
		req.Params.Append(n.wildcardName, path[i+1:])
	} else if i < len(path)-2 {
		ctx.WriteError(StdHttpErrors[http.StatusNotFound])
		return
	}
	n.handler.Handle(ctx)
}
