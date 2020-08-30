package graphqlx

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
	"testing"
)

func TestN(t *testing.T) {
	s := NewSchema()

	type Product struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Info  string  `json:"info,omitempty"`
		Price float64 `json:"price"`
		X     []int64
	}

	type QueryOneProductForm struct {
		ID int64 `validator:"id:"`
	}

	s.AddQuery(
		"product",
		"just query one product by id",
		NewPair(
			QueryOneProductForm{}, Product{},
			func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error) {
				q := in.(*QueryOneProductForm)
				return &Product{ID: q.ID, Name: "sss", Info: "sdsd", Price: 12.12, X: []int64{1, 3, 4}}, nil
			},
		),
	)
	type CreateOneProductForm struct {
		Name  string
		Info  string
		Price float64 `validator:"V<0-1000>"`
	}
	s.AddMutation(
		"create",
		"create a new product",
		NewPair(
			CreateOneProductForm{}, Product{},
			func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error) {
				q := in.(*CreateOneProductForm)
				fmt.Println(q)
				return &Product{ID: 17, Name: q.Name, Info: q.Info, Price: q.Price}, nil
			},
		),
	)

	// use custom scalar type C
	type UpdateProductForm struct {
		Id int64
		C
	}

	type UpdateStatus struct {
		Ok bool
		C
	}
	// 127.0.0.1:8080/product?query=mutation+_{update(id:17, cid:12, cname:"sadad"){ok,c}}
	s.AddMutation(
		"update",
		"update a product",
		NewPair(
			UpdateProductForm{}, UpdateStatus{},
			func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error) {
				q := in.(*UpdateProductForm)
				fmt.Println(q)
				return &UpdateStatus{Ok: true}, nil
			},
		),
	)

	root := router.New()
	root.GET("/product", s.MakeHttpHandler("query"))

	_ = fasthttp.ListenAndServe("127.0.0.1:8080", root.Handler)
}
