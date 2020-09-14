package graphqlx

import (
	"context"
	"fmt"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/router"
)

func TestN(_ *testing.T) {
	s := NewSchema()

	type Product struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Info  string  `json:"info,omitempty"`
		Price float64 `json:"price"`
	}

	type QueryOneProductForm struct {
		ID int64 `validator:"id:"`
	}

	s.AddQueryFromFunction(
		"product",
		"just query one product by id",
		func(ctx context.Context, in *QueryOneProductForm, info *graphql.ResolveInfo) (*Product, error) {
			return &Product{ID: in.ID}, nil
		},
	)
	type CreateOneProductForm struct {
		Name  string
		Info  string
		Price float64 `validator:"V<0-1000>"`
	}
	s.AddMutationFromFunction(
		"create",
		"create a new product",
		func(ctx context.Context, in *CreateOneProductForm, info *graphql.ResolveInfo) (out *Product, err error) {
			return &Product{ID: 17, Name: in.Name, Info: in.Info, Price: in.Price}, nil
		},
	)

	// use custom scalar type C
	type UpdateProductForm struct {
		C
		Id int64
	}

	type UpdateStatus struct {
		C  C
		Ok bool
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

	root := router.New(nil)
	root.GET("/product", s.MakeHttpHandler("query"))

	_ = fasthttp.ListenAndServe("127.0.0.1:8080", root.Handler)
}
