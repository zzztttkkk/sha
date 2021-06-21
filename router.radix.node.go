// copied from `https://github.com/fasthttp/router`

package sha

import (
	"sort"
	"strings"

	"github.com/zzztttkkk/sha/utils"
)

func newNode(path string) *_Node {
	return &_Node{
		nType: static,
		path:  path,
	}
}

// conflict raises a panic with some details
func (n *_NodeWildcard) conflict(path, fullPath string) error {
	prefix := fullPath[:strings.LastIndex(fullPath, path)] + n.path

	return newRadixError(errWildcardConflict, path, fullPath, n.path, prefix)
}

// wildPathConflict raises a panic with some details
func (n *_Node) wildPathConflict(path, fullPath string) error {
	pathSeg := strings.SplitN(path, "/", 2)[0]
	prefix := fullPath[:strings.LastIndex(fullPath, path)] + n.path

	return newRadixError(errWildPathConflict, pathSeg, fullPath, n.path, prefix)
}

// clone clones the current _Node in a new pointer
func (n _Node) clone() *_Node {
	cloneNode := new(_Node)
	cloneNode.nType = n.nType
	cloneNode.path = n.path
	cloneNode.tsr = n.tsr
	cloneNode.handler = n.handler

	if len(n.children) > 0 {
		cloneNode.children = make([]*_Node, len(n.children))

		for i, child := range n.children {
			cloneNode.children[i] = child.clone()
		}
	}

	if n.wildcard != nil {
		cloneNode.wildcard = &_NodeWildcard{
			path:     n.wildcard.path,
			paramKey: n.wildcard.paramKey,
			handler:  n.wildcard.handler,
		}
	}

	if len(n.paramKeys) > 0 {
		cloneNode.paramKeys = make([]string, len(n.paramKeys))
		copy(cloneNode.paramKeys, n.paramKeys)
	}

	cloneNode.paramRegex = n.paramRegex

	return cloneNode
}

func (n *_Node) split(i int) {
	cloneChild := n.clone()
	cloneChild.nType = static
	cloneChild.path = cloneChild.path[i:]
	cloneChild.paramKeys = nil
	cloneChild.paramRegex = nil

	n.path = n.path[:i]
	n.handler = nil
	n.tsr = false
	n.wildcard = nil
	n.children = append(n.children[:0], cloneChild)
}

func (n *_Node) findEndIndexAndValues(path string) (int, []string) {
	index := n.paramRegex.FindStringSubmatchIndex(path)
	if len(index) == 0 {
		return -1, nil
	}

	end := index[1]

	index = index[2:]
	values := make([]string, len(index)/2)

	i := 0
	for j := range index {
		if (j+1)%2 != 0 {
			continue
		}

		values[i] = path[index[j-1]:index[j]]

		i++
	}

	return end, values
}

type _AutoOptionsHandler struct {
	m  []byte
	cm string
}

func (a *_AutoOptionsHandler) Handle(ctx *RequestCtx) {
	ctx.Response.Header().Set(HeaderAllow, a.m)
}

func newAutoOptions(method string) *_AutoOptionsHandler {
	aoh := &_AutoOptionsHandler{
		m:  []byte(method),
		cm: method,
	}
	return aoh
}

func isAutoOptionsHandler(h RequestHandler) bool {
	_, ok := h.(*_AutoOptionsHandler)
	return ok
}

func (n *_Node) setHandler(handler RequestHandler, fullPath string) (*_Node, error) {
	if n.handler != nil || n.tsr {
		oAoh, oAohOk := n.handler.(*_AutoOptionsHandler)
		nAoh, nAohOk := handler.(*_AutoOptionsHandler)

		if !oAohOk && !nAohOk {
			return n, newRadixError(errSetHandler, fullPath)
		}

		if oAohOk && nAohOk {
			oAoh.m = append(oAoh.m, ',', ' ')
			oAoh.m = append(oAoh.m, nAoh.cm...)
			handler = n.handler
		} else if !oAohOk && nAohOk {
			handler = n.handler
			//lint:ignore SA9003 nothing
		} else {
			// oAohOK && !nAohOk
		}
	}

	n.handler = handler
	foundTSR := false

	// Set TSR in method
	for i := range n.children {
		child := n.children[i]

		if child.path != "/" {
			continue
		}

		child.tsr = true
		foundTSR = true

		break
	}

	if n.path != "/" && !foundTSR {
		childTSR := newNode("/")
		childTSR.tsr = true
		n.children = append(n.children, childTSR)
	}

	return n, nil
}

func (n *_Node) insert(path, fullPath string, handler RequestHandler) (*_Node, error) {
	end := segmentEndIndex(path, true)
	child := newNode(path)

	wp := findWildPath(path, fullPath)
	if wp != nil {
		j := end
		if wp.start > 0 {
			j = wp.start
		}

		child.path = path[:j]

		if wp.start > 0 {
			n.children = append(n.children, child)

			return child.insert(path[j:], fullPath, handler)
		}

		switch wp.pType {
		case param:
			n.hasWildChild = true

			child.nType = wp.pType
			child.paramKeys = wp.keys
			child.paramRegex = wp.regex
		case wildcard:
			if len(path) == end && n.path[len(n.path)-1] != '/' {
				return nil, newRadixError(errWildcardSlash, fullPath)
			} else if len(path) != end {
				return nil, newRadixError(errWildcardNotAtEnd, fullPath)
			}

			if n.path != "/" && n.path[len(n.path)-1] == '/' {
				n.split(len(n.path) - 1)
				n.tsr = true

				n = n.children[0]
			}

			if n.wildcard != nil {
				if n.wildcard.path == path {
					return n, newRadixError(errSetWildcardHandler, fullPath)
				}

				return nil, n.wildcard.conflict(path, fullPath)
			}

			n.wildcard = &_NodeWildcard{
				path:     wp.path,
				paramKey: wp.keys[0],
				handler:  handler,
			}

			return n, nil
		}

		path = path[wp.end:]

		if len(path) > 0 {
			n.children = append(n.children, child)

			return child.insert(path, fullPath, handler)
		}
	}

	child.handler = handler
	n.children = append(n.children, child)

	if child.path == "/" {
		// Add TSR when split a edge and the remain path to insert is "/"
		n.tsr = true
	} else if strings.HasSuffix(child.path, "/") {
		child.split(len(child.path) - 1)
		child.tsr = true
	} else {
		childTSR := newNode("/")
		childTSR.tsr = true
		child.children = append(child.children, childTSR)
	}

	return child, nil
}

// add adds the handler to _Node for the given path
func (n *_Node) add(path, fullPath string, handler RequestHandler) (*_Node, error) {
	if len(path) == 0 {
		return n.setHandler(handler, fullPath)
	}

	for _, child := range n.children {
		i := longestCommonPrefix(path, child.path)
		if i == 0 {
			continue
		}

		switch child.nType {
		case static:
			if len(child.path) > i {
				child.split(i)
			}

			if len(path) > i {
				return child.add(path[i:], fullPath, handler)
			}
		case param:
			wp := findWildPath(path, fullPath)

			isParam := wp.start == 0 && wp.pType == param
			hasHandler := child.handler != nil || handler == nil

			if len(path) == wp.end && isParam && hasHandler {
				// The current segment is a param and it's duplicated
				if child.path == path {
					return child, newRadixError(errSetHandler, fullPath)
				}

				return nil, child.wildPathConflict(path, fullPath)
			}

			if len(path) > i {
				if child.path == wp.path {
					return child.add(path[i:], fullPath, handler)
				}

				return n.insert(path, fullPath, handler)
			}
		}

		if path == "/" {
			n.tsr = true
		}

		return child.setHandler(handler, fullPath)
	}

	return n.insert(path, fullPath, handler)
}

func (n *_Node) getFromChild(path string, ctx *RequestCtx) (RequestHandler, bool) {
	var parent *_Node

	parentIndex, childIndex := 0, 0

walk:
	for {
		for _, child := range n.children[childIndex:] {
			childIndex++

			switch child.nType {
			case static:

				// Checks if the first byte is equal
				// It's faster than compare strings
				if path[0] != child.path[0] {
					continue
				}

				if len(path) > len(child.path) {
					if path[:len(child.path)] != child.path {
						continue
					}

					path = path[len(child.path):]

					parent = n
					n = child

					parentIndex = childIndex
					childIndex = 0

					continue walk

				} else if path == child.path {
					switch {
					case child.tsr:
						return nil, true
					case child.handler != nil:
						return child.handler, false
					case child.wildcard != nil:
						if ctx != nil {
							ctx.Request.URL.Params.Set(child.wildcard.paramKey, utils.B(""))
						}

						return child.wildcard.handler, false
					}

					return nil, false
				}

			case param:
				end := segmentEndIndex(path, false)
				values := []string{copyString(path[:end])}

				if child.paramRegex != nil {
					end, values = child.findEndIndexAndValues(path[:end])
					if end == -1 {
						continue
					}
				}

				if len(path) > end {
					h, tsr := child.getFromChild(path[end:], ctx)
					if tsr {
						return nil, tsr
					} else if h != nil {
						if ctx != nil {
							for i, key := range child.paramKeys {
								ctx.Request.URL.Params.Set(key, utils.B(values[i]))
							}
						}

						return h, false
					}

				} else if len(path) == end {
					switch {
					case child.tsr:
						return nil, true
					case child.handler == nil:
						// try another child
						continue
					case ctx != nil:
						for i, key := range child.paramKeys {
							ctx.Request.URL.Params.Set(key, utils.B(values[i]))
						}
					}

					return child.handler, false
				}

			default:
				panic("invalid node type")
			}

		}

		// Go back and continue with the remaining children of the parent
		// to try to discover the correct child node
		// if the parent has a child node of type param
		//
		// See: https://github.com/fasthttp/router/issues/37
		if parent != nil && parent.hasWildChild && len(parent.children[parentIndex:]) > 0 {
			path = n.path + path
			childIndex = parentIndex

			n = parent
			parent = nil

			continue walk
		}

		if n.wildcard != nil {
			if ctx != nil {
				ctx.Request.URL.Params.Set(n.wildcard.paramKey, utils.B(path))
			}

			return n.wildcard.handler, false
		}

		return nil, false
	}
}

func (n *_Node) find(path string, buf *utils.Buf) (bool, bool) {
	if len(path) > len(n.path) {
		if !strings.EqualFold(path[:len(n.path)], n.path) {
			return false, false
		}

		path = path[len(n.path):]
		buf.WriteString(n.path)

		found, tsr := n.findFromChild(path, buf)
		if found {
			return found, tsr
		}

		bufferRemoveString(buf, n.path)

	} else if strings.EqualFold(path, n.path) {
		buf.WriteString(n.path)

		if n.tsr {
			if n.path == "/" {
				bufferRemoveString(buf, n.path)
			} else {
				_ = buf.WriteByte('/')
			}

			return true, true
		}

		return n.handler != nil, false
	}

	return false, false
}

func (n *_Node) findFromChild(path string, buf *utils.Buf) (bool, bool) {
	for _, child := range n.children {
		switch child.nType {
		case static:
			found, tsr := child.find(path, buf)
			if found {
				return found, tsr
			}

		case param:
			end := segmentEndIndex(path, false)

			if child.paramRegex != nil {
				end, _ = child.findEndIndexAndValues(path[:end])
				if end == -1 {
					continue
				}
			}

			buf.WriteString(path[:end])

			if len(path) > end {
				found, tsr := child.findFromChild(path[end:], buf)
				if found {
					return found, tsr
				}

			} else if len(path) == end {
				if child.tsr {
					_ = buf.WriteByte('/')

					return true, true
				}

				return child.handler != nil, false
			}

			bufferRemoveString(buf, path[:end])

		default:
			panic("invalid node type")
		}
	}

	if n.wildcard != nil {
		buf.WriteString(path)

		return true, false
	}

	return false, false
}

// sort sorts the current _Node and their children
func (n *_Node) sort() {
	for _, child := range n.children {
		child.sort()
	}

	sort.Sort(n)
}

// Len returns the total number of children the node has
func (n *_Node) Len() int {
	return len(n.children)
}

// Swap swaps the order of children nodes
func (n *_Node) Swap(i, j int) {
	n.children[i], n.children[j] = n.children[j], n.children[i]
}

// Less checks if the node 'i' has less priority than the node 'j'
func (n *_Node) Less(i, j int) bool {
	if n.children[i].nType < n.children[j].nType {
		return true
	} else if n.children[i].nType > n.children[j].nType {
		return false
	}

	return len(n.children[i].children) > len(n.children[j].children)
}
