// copied from `https://github.com/fasthttp/router`

package sha

import (
	"github.com/zzztttkkk/sha/utils"
	"strings"
)

// New returns an empty routes storage
func newRadixTree() *_RadixTree {
	return &_RadixTree{
		root: &_Node{
			nType: root,
		},
	}
}

// HTTPItem adds a node with the given handle to the path.
//
// WARNING: Not concurrency-safe!
//goland:noinspection GoNilness
func (t *_RadixTree) Add(path string, handler RequestHandler) {
	if !strings.HasPrefix(path, "/") {
		panicf("path must begin with '/' in path '%s'", path)
	} else if handler == nil {
		panic("nil handler")
	}

	fullPath := path

	i := longestCommonPrefix(path, t.root.path)
	if i > 0 {
		if len(t.root.path) > i {
			t.root.split(i)
		}

		path = path[i:]
	}

	n, err := t.root.add(path, fullPath, handler)
	if err != nil {
		radixErr := err.(*radixError)
		if t.Mutable && !n.tsr {
			switch radixErr.msg {
			case errSetHandler:
				n.handler = handler
				return
			case errSetWildcardHandler:
				n.wildcard.handler = handler
				return
			}
		}

		panic(err)
	}

	if len(t.root.path) == 0 {
		t.root = t.root.children[0]
		t.root.nType = root
	}

	// Reorder the nodes
	t.root.sort()
}

// Get returns the handle registered with the given path (key). The values of
// param/wildcard are saved as ctx.UserValue.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (t *_RadixTree) Get(path string, ctx *RequestCtx) (RequestHandler, bool) {
	if len(path) > len(t.root.path) {
		if path[:len(t.root.path)] != t.root.path {
			return nil, false
		}

		path = path[len(t.root.path):]

		return t.root.getFromChild(path, ctx)

	} else if path == t.root.path {
		switch {
		case t.root.tsr:
			return nil, true
		case t.root.handler != nil:
			return t.root.handler, false
		case t.root.wildcard != nil:
			if ctx != nil {
				ctx.Request.URL.Params.SetString(t.root.wildcard.paramKey, "")
			}
			return t.root.wildcard.handler, false
		}
	}

	return nil, false
}

// FindCaseInsensitivePath makes a case-insensitive lookup of the given path
// and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (t *_RadixTree) FindCaseInsensitivePath(path string, fixTrailingSlash bool, buf *utils.Buf) bool {
	found, tsr := t.root.find(path, buf)

	if !found || (tsr && !fixTrailingSlash) {
		buf.Reset()

		return false
	}

	return true
}
