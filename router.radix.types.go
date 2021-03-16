// copied from `https://github.com/fasthttp/router`

package sha

import (
	"regexp"
)

type _NodeType uint8

type _NodeWildcard struct {
	path     string
	paramKey string
	handler  RequestHandler
}

type _Node struct {
	nType _NodeType

	path         string
	tsr          bool
	handler      RequestHandler
	hasWildChild bool
	children     []*_Node
	wildcard     *_NodeWildcard

	paramKeys  []string
	paramRegex *regexp.Regexp
}

type _WildPath struct {
	path  string
	keys  []string
	start int
	end   int
	pType _NodeType

	pattern string
	regex   *regexp.Regexp
}

// _RadixTree is a routes storage
type _RadixTree struct {
	root *_Node

	// If enabled, the node handler could be updated
	Mutable bool
}
