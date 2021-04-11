package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/validator"
	"strings"
)

type _DocForm struct {
	Prefix    string   `validator:"prefix,desc='url prefix',optional"`
	Tags      []string `validator:"tags,optional"`
	TagsLogic string   `validator:"logic,optional"`
}

func (d _DocForm) DefaultTagsLogic() interface{} { return "OR" }

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

func (m *Mux) ServeDocument(method, path string, middleware ...Middleware) {
	m.HTTPWithOptions(
		&HandlerOptions{Document: validator.NewDocument(_DocForm{}, validator.Undefined), Middlewares: middleware},
		method, path,
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form _DocForm
			ctx.MustValidateForm(&form)
			ctx.Response.Header().SetContentType(MIMEMarkdown)

			pm := map[string]map[string]validator.Document{}
			if len(form.Prefix) > 0 {
				for p, m1 := range m.documents {
					if strings.HasPrefix(p, form.Prefix) {
						pm[p] = m1
					}
				}
			} else {
				pm = m.documents
			}

			var vm map[string]map[string]validator.Document
			if len(form.Tags) > 0 {
				vm = map[string]map[string]validator.Document{}
				switch form.TagsLogic {
				case "OR":
					for _, tag := range form.Tags {
						tag = strings.ToLower(tag)
						for p1, m1 := range pm {
							for m2, d := range m1 {
								if contains(d.Tags(), tag) {
									mapAppend(vm, p1, m2, d)
								}
							}
						}
					}
				default:
					for p1, m1 := range pm {
						for m2, d := range m1 {
							for _, tag := range form.Tags {
								if !contains(d.Tags(), tag) {
									break
								}
								mapAppend(vm, p1, m2, d)
							}
						}
					}
				}
			} else {
				vm = pm
			}

			var buf strings.Builder

			for p, m1 := range vm {
				buf.WriteString(fmt.Sprintf("## Path: %s\n", p))
				for me, doc := range m1 {
					buf.WriteString(fmt.Sprintf("### Method: %s\n", me))
					buf.WriteString(fmt.Sprintf(
						`## Document:
#### Input:
%s
#### Output:
%s
#### Description:
%s
#### Tags:
%s
`,
						doc.Input(), doc.Output(), doc.Description(), strings.Join(doc.Tags(), "; "),
					))
				}
			}

			_, _ = ctx.WriteString(buf.String())
		}),
	)
}
