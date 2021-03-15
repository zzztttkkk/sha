// copied from `https://github.com/fasthttp/router`

package sha

import (
	"regexp"
)

type nodeType uint8

type nodeWildcard struct {
	path     string
	paramKey string
	handler  RequestHandler
}

type node struct {
	nType nodeType

	path         string
	tsr          bool
	handler      RequestHandler
	hasWildChild bool
	children     []*node
	wildcard     *nodeWildcard

	paramKeys  []string
	paramRegex *regexp.Regexp
}

type wildPath struct {
	path  string
	keys  []string
	start int
	end   int
	pType nodeType

	pattern string
	regex   *regexp.Regexp
}

// _RadixTree is a routes storage
type _RadixTree struct {
	root *node

	// If enabled, the node handler could be updated
	Mutable bool
}
