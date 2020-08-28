package graphqlx

import (
	"github.com/graphql-go/graphql"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/output"
	"strings"
)

type Schema struct {
	qs graphql.Fields
	ms graphql.Fields
}

func NewSchema() *Schema {
	return &Schema{
		qs: graphql.Fields{},
		ms: graphql.Fields{},
	}
}

func (s *Schema) AddQuery(name, descp string, pair *Pair) *Schema {
	field := pair.toField(name, descp)
	s.qs[name] = field
	return s
}

func (s *Schema) AddMutation(name, descp string, pair *Pair) *Schema {
	field := pair.toField(name, descp)
	s.ms[name] = field
	return s
}

func (s *Schema) toGraphqlSchema() (graphql.Schema, error) {
	sc := graphql.SchemaConfig{}
	if len(s.qs) > 0 {
		sc.Query = graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Query",
				Fields: s.qs,
			},
		)
	}

	if len(s.ms) > 0 {
		sc.Mutation = graphql.NewObject(
			graphql.ObjectConfig{
				Name:   "Mutation",
				Fields: s.ms,
			},
		)
	}

	return graphql.NewSchema(sc)
}

func (s *Schema) MakeHttpHandler(formName string) fasthttp.RequestHandler {
	schema, err := s.toGraphqlSchema()
	if err != nil {
		panic(err)
	}

	return func(ctx *fasthttp.RequestCtx) {
		query := ctx.FormValue(formName)
		if len(query) < 1 {
			output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
			return
		}

		result := graphql.Do(
			graphql.Params{
				Context:       ctxs.Wrap(ctx),
				Schema:        schema,
				RequestString: gotils.B2S(query),
			},
		)

		if len(result.Errors) > 0 {
			var err error
			buf := strings.Builder{}
			end := len(result.Errors) - 1
			for i, e := range result.Errors {
				buf.WriteString(e.Error())
				if i < end {
					buf.WriteByte(';')
				}
			}
			err = output.NewError(fasthttp.StatusBadRequest, 0, buf.String())
			output.Error(ctx, err)
			return
		}
		output.MsgOK(ctx, result.Data)
	}
}
