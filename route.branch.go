package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
	"reflect"
	"strings"
)

type _RouteBranch struct {
	_MiddlewareNode
	allHandlers map[string]map[string]RequestHandler

	root         *Mux
	parentRouter *_RouteBranch
	prefix       string
	children     map[string]*_RouteBranch
}

var _ Router = (*_RouteBranch)(nil)

func (branch *_RouteBranch) HTTP(method, path string, handler RequestHandler) {
	method = strings.ToUpper(method)
	m := branch.allHandlers[method]
	if m == nil {
		m = map[string]RequestHandler{}
		branch.allHandlers[method] = m
	}
	m[path] = handler
}

func (branch *_RouteBranch) WebSocket(path string, wh WebSocketHandlerFunc) {
	branch.HTTP("get", path, wshToHandler(wh))
}

func (branch *_RouteBranch) HTTPWithForm(method, path string, handler RequestHandler, form interface{}) {
	if form == nil {
		branch.HTTP(method, path, handler)
		return
	}

	branch.HTTP(
		method, path,
		&_FormRequestHandler{
			RequestHandler: handler,
			Documenter:     validator.GetRules(reflect.TypeOf(form)),
		},
	)
}

func (branch *_RouteBranch) FileSystem(fs http.FileSystem, method, path string, autoIndex bool, middleware ...Middleware) {
	branch.HTTP(
		method, path,
		fileSystemHandler(fs, path, autoIndex, middleware...),
	)
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
		for p, h := range b {
			branch.goUp(a, p, branch.wrap(h))
		}
	}
}

func (branch *_RouteBranch) goUp(method, path string, handler RequestHandler) {
	if branch.parentRouter != nil {
		branch.parentRouter.goUp(method, branch.prefix+path, handler)
	} else {
		branch.root.doAddHandler(method, branch.prefix+path, handler)
	}
}

func NewBranch() Router {
	return &_RouteBranch{
		allHandlers: map[string]map[string]RequestHandler{},
		children:    map[string]*_RouteBranch{},
	}
}
