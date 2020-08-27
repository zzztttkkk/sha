package grfqlx

import (
	"context"
	"github.com/graphql-go/graphql"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"testing"
)

type Product struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Info  string  `json:"info,omitempty"`
	Price float64 `json:"price"`
	X     []int64
}

type QueryOneProduct struct {
	ID int64 `validator:"id:"`
}

func TestN(t *testing.T) {
	_ShowFmtError = true

	s := NewSchema()

	s.AddQuery(
		"product",
		"just query one product by id",
		NewPair(
			QueryOneProduct{}, Product{},
			func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error) {
				q := in.(*QueryOneProduct)
				return &Product{ID: q.ID, Name: "sss", Info: "sdsd", Price: 12.12, X: []int64{1, 3, 4}}, nil
			},
		),
	)

	root := router.New()
	root.GET("/product", s.NewHandler("query"))

	_ = fasthttp.ListenAndServe("127.0.0.1:8080", root.Handler)
}
