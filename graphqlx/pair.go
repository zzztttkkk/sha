package graphqlx

import (
	"context"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/zzztttkkk/suna/validator"
)

type ResolveFunc func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error)

type Pair struct {
	in    interface{}
	rules *validator.Rules
	am    graphql.FieldConfigArgument

	out     interface{}
	resolve ResolveFunc
}

func NewPair(inV, outV interface{}, resolveFunc ResolveFunc) *Pair {
	p := &Pair{in: inV, out: outV}
	p.rules = validator.GetRules(inV)
	p.am = p.rules.ArgumentMap()

	p.resolve = resolveFunc
	return p
}

func (p *Pair) toField(name, descp string) *graphql.Field {
	var otype graphql.Output
	if p.out == nil {
		otype = nil
	} else {
		otype = NewOutObjectType(reflect.ValueOf(p.out))
	}

	field := &graphql.Field{
		Type:        otype,
		Name:        name,
		Description: descp,
		Args:        p.am,
	}
	field.Resolve = func(params graphql.ResolveParams) (interface{}, error) {
		v, err := p.rules.ValidateAndBind(params.Args)
		if err != nil {
			return nil, err
		}
		return p.resolve(params.Context, v.Addr().Interface(), &params.Info)
	}
	return field
}

func NewPairFromFunction(v interface{}) *Pair {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("suna.graphqlx: `%v` is not a function", v))
	}

	inT := t.In(1).Elem()
	outT := t.Out(0).Elem()
	fnV := reflect.ValueOf(v)

	return NewPair(
		reflect.New(inT).Elem().Interface(),
		reflect.New(outT).Elem().Interface(),
		func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error) {
			rspV := fnV.Call(
				[]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in), reflect.ValueOf(info)},
			)
			ev := rspV[1].Interface()
			if ev != nil {
				return nil, ev.(error)
			}
			return rspV[0].Interface(), nil
		},
	)
}
