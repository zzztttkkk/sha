package sha

type Documenter interface {
	Document() string
}

// request Handler with document
type DocedRequestHandler interface {
	Documenter
	RequestHandler
}

type Router interface {
	HTTP(method, path string, handler RequestHandler)
	WebSocket(path string, handler WebSocketHandlerFunc)
	RESTWithForm(method, path string, handler RequestHandler, form interface{})
	AddBranch(prefix string, router Router)
	Use(middleware ...Middleware)
}
