package utils

import (
	"regexp"
	"strings"
)

type M map[string]string

var placeholderRegexp = regexp.MustCompile("\\${.*?}")

type _SNode struct {
	str  string
	name string
}

func parseFs(fs string) []*_SNode {
	var nodes []*_SNode

	prev := 0
	for _, v := range placeholderRegexp.FindAllSubmatchIndex(B(fs), -1) {
		begin, end := v[0], v[1]
		nodes = append(nodes, &_SNode{str: fs[prev:begin]})
		prev = end
		txt := fs[begin+2 : end-1]
		nodes = append(nodes, &_SNode{name: txt})
		continue
	}
	nodes = append(nodes, &_SNode{str: fs[prev:]})
	return nodes
}

type NamedFmt struct {
	nodes []*_SNode
	Args  []string
}

func NewNamedFmt(fs string) *NamedFmt {
	f := &NamedFmt{nodes: parseFs(fs)}

	for _, node := range f.nodes {
		if len(node.name) > 0 {
			f.Args = append(f.Args, node.name)
		}
	}

	return f
}

func (f *NamedFmt) Render(v M) string {
	buf := strings.Builder{}
	for _, node := range f.nodes {
		if len(node.name) > 0 {
			buf.WriteString(v[node.name])
		} else {
			buf.WriteString(node.str)
		}
	}
	return buf.String()
}
