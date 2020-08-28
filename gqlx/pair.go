package gqlx

import (
	"context"
	"github.com/graphql-go/graphql"
	"github.com/zzztttkkk/suna/validator"
	"reflect"
)

type ResolveFunc func(ctx context.Context, in interface{}, info *graphql.ResolveInfo) (out interface{}, err error)

type Pair struct {
	field *graphql.Field

	in    interface{}
	rules *validator.Rules
	am    graphql.FieldConfigArgument

	out     interface{}
	resolve ResolveFunc
}

func NewPair(inV, outV interface{}, resolveFunc ResolveFunc) *Pair {
	p := &Pair{in: inV, out: outV}

	if inV != nil {
		p.rules = validator.GetRules(inV)
		p.am = p.rules.ArgumentMap()
	} else {
		p.am = nil
	}

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
