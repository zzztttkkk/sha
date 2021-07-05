package sha

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zzztttkkk/sha/validator"
)

type _DocForm struct {
	Prefix    string   `vld:"prefix,desc='url prefix',optional"`
	Tags      []string `vld:"tags,optional"`
	TagsLogic string   `vld:"logic,optional"`
}

func (_DocForm) Default(name string) func() interface{} {
	switch name {
	case "TagsLogic":
		return func() interface{} { return "OR" }
	default:
		return nil
	}
}

func contains(v []string, d string) bool {
	for _, a := range v {
		if a == d {
			return true
		}
	}
	return false
}

func mapAppend(m map[string]map[string]validator.Document, path, method string, doc validator.Document) {
	m1 := m[path]
	if m1 == nil {
		m1 = map[string]validator.Document{}
		m[path] = m1
	}
	m1[method] = doc
}

func (m *Mux) ServeDocuments(method, path string, middleware ...Middleware) {
	m.HTTPWithOptions(
		&RouteOptions{Document: validator.NewDocument(_DocForm{}, nil), Middlewares: middleware},
		method, path,
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form _DocForm
			ctx.MustValidateForm(&form)
			ctx.Response.Header().SetContentType(MIMEMarkdown)

			pathFilterMap := map[string]map[string]validator.Document{}
			if len(form.Prefix) > 0 {
				for p, m1 := range m.documents {
					if strings.HasPrefix(p, form.Prefix) {
						pathFilterMap[p] = m1
					}
				}
			} else {
				pathFilterMap = m.documents
			}

			var tagsFilterMap map[string]map[string]validator.Document
			if len(form.Tags) > 0 {
				tagsFilterMap = map[string]map[string]validator.Document{}
				switch form.TagsLogic {
				case "OR":
					for _, tag := range form.Tags {
						tag = strings.ToLower(tag)
						for p1, m1 := range pathFilterMap {
							for m2, d := range m1 {
								if contains(d.Tags(), tag) {
									mapAppend(tagsFilterMap, p1, m2, d)
								}
							}
						}
					}
				default:
					for p1, m1 := range pathFilterMap {
						for m2, d := range m1 {
							for _, tag := range form.Tags {
								if !contains(d.Tags(), tag) {
									break
								}
								mapAppend(tagsFilterMap, p1, m2, d)
							}
						}
					}
				}
			} else {
				tagsFilterMap = pathFilterMap
			}

			var buf strings.Builder
			type _PathItem struct {
				Path string
				doc  string
			}
			var paths []*_PathItem

			for p, m1 := range tagsFilterMap {
				buf.WriteString(fmt.Sprintf("## Path: %s\n", p))
				for me, doc := range m1 {
					buf.WriteString(fmt.Sprintf("### Method: %s\n", me))
					if doc.Description() != "" {
						buf.WriteString(fmt.Sprintf("#### Description:\r\n%s\r\n", doc.Input()))
					}
					if doc.Input() != "" {
						buf.WriteString(fmt.Sprintf("#### Input:\r\n%s\r\n", doc.Input()))
					}
					if doc.Output() != "" {
						buf.WriteString(fmt.Sprintf("#### Output:\r\n%s\r\n", doc.Output()))
					}
					if len(doc.Tags()) > 0 {
						buf.WriteString(fmt.Sprintf("#### Tags:\r\n%s\r\n", strings.Join(doc.Tags(), "; ")))
					}
				}
				paths = append(paths, &_PathItem{doc: buf.String(), Path: p})
				buf.Reset()
			}

			sort.Slice(paths, func(i, j int) bool { return paths[i].Path < paths[j].Path })

			for _, v := range paths {
				buf.WriteString(v.doc)
			}
			_ = ctx.WriteString(buf.String())
		}),
	)
}
