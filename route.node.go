package suna

import (
	"fmt"
	"github.com/zzztttkkk/suna/internal"
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

	nl []*_RouteNode

	methods     []byte
	autoHandler bool
}

func (node *_RouteNode) addMethod(method string) {
	if len(node.methods) > 0 {
		node.methods = append(node.methods, ',', ' ')
	}
	node.methods = append(node.methods, []byte(method)...)
}

func (node *_RouteNode) handleOptions(ctx *RequestCtx) {
	ctx.Response.Header.Set(internal.B(HeaderAllow), node.methods)
}

func (node *_RouteNode) addChild(c *_RouteNode) {
	node.nl = append(node.nl, c)

}

func (node *_RouteNode) getChild(name string) *_RouteNode {
	for _, n := range node.nl {
		if n.name == name {
			return n
		}
	}
	return nil
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

		sn := node.getChild(p)
		if sn == nil {
			sn = &_RouteNode{name: p, parent: node}
			node.addChild(sn)
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

	if len(node.nl) != 0 || len(node.params) != 0 || len(node.wildcardName) != 0 {
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
	for _, n := range node.nl {
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

			if len(n.nl) < 1 {
				return prevI, nil
			}

			n = n.getChild(internal.S(key))
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

	if len(n.nl) < 1 {
		return prevI, nil
	}

	n = n.getChild(internal.S(key))
	if n == nil {
		return prevI, nil
	}
	return i, n
}
