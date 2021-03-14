package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
	"strings"
)

type _BranchNode struct {
	h RequestHandler
	d validator.Document
}

type _RouteBranch struct {
	_MiddlewareNode
	allHandlers map[string]map[string]_BranchNode

	root         *Mux
	parentRouter *_RouteBranch
	prefix       string
	children     map[string]*_RouteBranch
}

var _ Router = (*_RouteBranch)(nil)

func (branch *_RouteBranch) HTTPWithDocument(method, path string, handler RequestHandler, doc validator.Document) {
	method = strings.ToUpper(method)
	m := branch.allHandlers[method]
	if m == nil {
		m = map[string]_BranchNode{}
		branch.allHandlers[method] = m
	}
	m[path] = _BranchNode{h: handler, d: doc}
}

func (branch *_RouteBranch) HTTP(method, path string, handler RequestHandler) {
	branch.HTTPWithDocument(method, path, handler, nil)
}

func (branch *_RouteBranch) HTTPWithMiddleware(middleware []Middleware, method, path string, handler RequestHandler) {
	branch.HTTP(method, path, handlerWithMiddleware(handler, middleware...))
}

func (branch *_RouteBranch) HTTPWithMiddlewareAndDocument(middleware []Middleware, method, path string, handler RequestHandler, doc validator.Document) {
	branch.HTTPWithDocument(method, path, handlerWithMiddleware(handler, middleware...), doc)
}

func (branch *_RouteBranch) WebSocket(path string, wh WebSocketHandlerFunc) {
	branch.HTTP("get", path, wshToHandler(wh))
}

func (branch *_RouteBranch) FilePath(fs http.FileSystem, method, path string, autoIndex bool, middleware ...Middleware) {
	branch.HTTP(
		method, path,
		makeFileSystemHandler(fs, path, autoIndex, middleware...),
	)
}

func (branch *_RouteBranch) File(fs http.FileSystem, filename, method, path string, middleware ...Middleware) {
	branch.HTTP(method, path, makeFileHandler(fs, filename, middleware...))
}

func (branch *_RouteBranch) AddBranch(prefix string, router Router) {
	v, ok := router.(*_RouteBranch)
	if !ok {
		panic(fmt.Errorf("sha.router: `%v` is not a branch", router))
	}
	branch.children[prefix] = v
	v.prefix = prefix
	v._MiddlewareNode.parentMwNode = &branch._MiddlewareNode
	v.parentRouter = branch
}

func (branch *_RouteBranch) goDown() {
	for _, v := range branch.children {
		v.goDown()
	}

	for a, b := range branch.allHandlers {
		for p, n := range b {
			branch.goUp(a, p, branch.wrap(n.h), n.d)
		}
	}
}

func (branch *_RouteBranch) goUp(method, path string, handler RequestHandler, doc validator.Document) {
	if branch.parentRouter != nil {
		branch.parentRouter.goUp(method, branch.prefix+path, handler, doc)
	} else {
		branch.root.doAddHandler(method, branch.prefix+path, handler, false, doc)
	}
}

func NewBranch() Router {
	return &_RouteBranch{
		allHandlers: map[string]map[string]_BranchNode{},
		children:    map[string]*_RouteBranch{},
	}
}
