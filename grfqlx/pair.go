package grfqlx

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
	p.rules = validator.GetRules(inV)
	p.am = p.rules.ArgumentMap()
	p.resolve = resolveFunc
	return p
}

func (p *Pair) toField(name, descp string) *graphql.Field {
	field := &graphql.Field{
		Type:        NewOutObjectType(reflect.ValueOf(p.out)),
		Name:        name,
		Description: descp,
		Args:        p.am,
	}
	field.Resolve = func(params graphql.ResolveParams) (interface{}, error) {
		ele := p.rules.BindValue(params.Args).Addr().Interface()
		return p.resolve(params.Context, ele, &params.Info)
	}
	return field
}
