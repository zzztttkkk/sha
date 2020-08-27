package validator

import (
	"encoding/json"
	"errors"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/zzztttkkk/suna/jsonx"
)

func _JsonObjSerialize(m jsonx.Object) string {
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func _JsonArySerialize(m jsonx.Array) string {
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(data)
}

var JsonObjectGraphqlType = graphql.NewScalar(
	graphql.ScalarConfig{
		Name: "Object",
		Serialize: func(value interface{}) interface{} {
			switch rv := value.(type) {
			case jsonx.Object:
				return _JsonObjSerialize(rv)
			case *jsonx.Object:
				if rv == nil {
					return nil
				}
				return _JsonObjSerialize(*rv)
			case map[string]interface{}:
				return _JsonObjSerialize(rv)
			case *map[string]interface{}:
				if rv == nil {
					return nil
				}
				return _JsonObjSerialize(*rv)
			default:
				panic(errors.New("suna.validator: not a json object"))
				return nil
			}
		},
		ParseValue: func(value interface{}) interface{} {
			m, e := jsonx.ParseObject(value)
			if e != nil {
				panic(e)
			}
			return m
		},
		ParseLiteral: func(value ast.Value) interface{} {
			switch rv := value.(type) {
			case *ast.StringValue:
				m, e := jsonx.ParseObject(rv.Value)
				if e != nil {
					panic(e)
				}
				return m
			default:
				return nil
			}
		},
	},
)

var JsonArrayGraphqlType = graphql.NewScalar(
	graphql.ScalarConfig{
		Name: "Array",
		Serialize: func(value interface{}) interface{} {
			switch rv := value.(type) {
			case jsonx.Array:
				return _JsonArySerialize(rv)
			case *jsonx.Array:
				if rv == nil {
					return nil
				}
				return _JsonArySerialize(*rv)
			case []interface{}:
				return _JsonArySerialize(rv)
			case *[]interface{}:
				if rv == nil {
					return nil
				}
				return _JsonArySerialize(*rv)
			default:
				panic(errors.New("suna.validator: not a json array"))
				return nil
			}
		},
		ParseValue: func(value interface{}) interface{} {
			m, e := jsonx.ParseArray(value)
			if e != nil {
				panic(e)
			}
			return m
		},
		ParseLiteral: func(value ast.Value) interface{} {
			switch rv := value.(type) {
			case *ast.StringValue:
				m, e := jsonx.ParseArray(rv.Value)
				if e != nil {
					panic(e)
				}
				return m
			default:
				return nil
			}
		},
	},
)
