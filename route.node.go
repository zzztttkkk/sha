package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"sort"
	"strings"
)

type _RouteNode struct {
	param        []byte
	raw          string
	wildcardName string
	handler      RequestHandler
	parent       *_RouteNode
	name         string

	sChild      *_RouteNode
	children    []*_RouteNode
	childrenMap map[string]*_RouteNode

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
	ctx.Response.Header.Set(HeaderAllow, node.methods)
}

func (node *_RouteNode) addChild(c *_RouteNode) {
	node.children = append(node.children, c)
	c.parent = node
	sort.Slice(node.children, func(i, j int) bool { return node.children[i].name < node.children[j].name })

	if node.childrenMap == nil {
		node.childrenMap = map[string]*_RouteNode{}
	}
	node.childrenMap[c.name] = c
}

func (node *_RouteNode) getChild(name string) *_RouteNode {
	i, j := 0, len(node.children)
	if j < 1 {
		return nil
	}

	if j < 6 {
		for _, n := range node.children {
			if n.name == name {
				return n
			}
		}
		return nil
	}

	if j > 16 {
		return node.childrenMap[name]
	}

	for i < j {
		h := int(uint(i+j) >> 1)
		n := node.children[h]
		if n.name == name {
			return n
		}
		if n.name < name {
			i = h + 1
		} else {
			j = h
		}
	}
	return nil
}

func (node *_RouteNode) _getChild(name []byte) *_RouteNode {
	if node.sChild != nil {
		return node.sChild
	}
	return node.getChild(utils.S(name))
}

func (node *_RouteNode) addHandler(path []string, handler RequestHandler, raw string) *_RouteNode {
	if len(path) < 1 {
		if handler == nil {
			return node
		}

		if node.handler != nil {
			if node.autoHandler {
				node.handler = handler
				node.autoHandler = false
			} else {
				panic(fmt.Errorf("sha.router: `/%s` conflict with `%s`", raw, node.raw))
			}
			return node
		}

		node.handler = handler
		node.raw = raw
		return node
	}

	p := path[0]
	ind := strings.IndexByte(p, ':')
	if ind < 0 { // normal part
		if node.sChild != nil {
			panic(fmt.Errorf("sha.router: `/%s` conflict with others", raw))
		}

		sn := node.getChild(p)
		if sn == nil {
			sn = &_RouteNode{name: p, parent: node}
			node.addChild(sn)
		}
		return sn.addHandler(path[1:], handler, raw)
	}

	if ind == 0 { // param part
		if len(node.children) != 0 {
			panic(fmt.Errorf("sha.router: `/%s` conflict with others", raw))
		}
		node.sChild = &_RouteNode{param: []byte(p[1:])}
		return node.sChild.addHandler(path[1:], handler, raw)
	}

	if len(node.children) != 0 {
		panic(fmt.Errorf("sha.router: `/%s` conflict with others", raw))
	}
	if !strings.HasSuffix(p, ":*") || len(path) != 1 {
		panic(fmt.Errorf("sha.router: bad path value `/%s`", raw))
	}
	node.sChild = &_RouteNode{wildcardName: p[:len(p)-2]}
	return node.sChild.addHandler(nil, handler, raw)
}

func (node *_RouteNode) find(path []byte, kvs *utils.Kvs) (int, *_RouteNode) {
	n := node
	var temp []byte
	var i int
	var b byte
	for i, b = range path {
		if b == '/' {
			n = n._getChild(temp)
			if n == nil {
				return 0, nil
			}
			if len(n.param) > 0 {
				kvs.AppendBytes(n.param, temp)
				temp = temp[:0]
				continue
			}
			if len(n.wildcardName) > 0 {
				return i - len(temp), n
			}
			temp = temp[:0]
			continue
		}
		temp = append(temp, b)
	}
	n = n._getChild(temp)
	if n == nil {
		return 0, nil
	}
	if len(n.param) > 0 {
		if len(temp) < 1 {
			return 0, nil
		}
		kvs.AppendBytes(n.param, temp)
	}
	if len(n.wildcardName) > 0 {
		return i - len(temp), n
	}
	return i, n
}
