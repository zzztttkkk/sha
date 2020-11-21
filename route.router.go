package suna

type Documenter interface {
	Document() string
}

// request handler with document
type DocedRequestHandler interface {
	Documenter
	RequestHandler
}

type Router interface {
	AddHandler(method, path string, handler RequestHandler)
	AddHandlerWithForm(method, path string, handler RequestHandler, form interface{})
	AddBranch(prefix string, router Router)
	Use(middleware ...Middleware)
}
