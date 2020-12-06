package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/rbac/internal"
	"html/template"
)

func static(fn func()) {
	internal.Dig.Append(func(_ Router, _ _PermOK) { fn() })
}

func init() {
	var index *template.Template

	static(
		func() {
			var err error
			v, err := box.FindString("index.html")
			if err != nil {
				panic(err)
			}
			index, err = template.New("").Parse(v)
			if err != nil {
				panic(err)
			}
		},
	)

	register(
		"GET",
		"/",
		func(rctx context.Context, rw ReqWriter) {
			MustGrantedAll(rctx, PermRbacLogin)

			rw.WriteTemplate(index, map[string]interface{}{"Title": "Rbac"})
		},
		nil,
	)
}
