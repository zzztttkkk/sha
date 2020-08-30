package jsonx

import (
	"errors"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func _ObjSerialize(m Object) string {
	data, err := Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func _ArySerialize(m Array) string {
	data, err := Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(data)
}

var GraphqlObject = graphql.NewScalar(
	graphql.ScalarConfig{
		Name: "Object",
		Serialize: func(value interface{}) interface{} {
			switch rv := value.(type) {
			case Object:
				return _ObjSerialize(rv)
			case *Object:
				if rv == nil {
					return nil
				}
				return _ObjSerialize(*rv)
			case map[string]interface{}:
				return _ObjSerialize(rv)
			case *map[string]interface{}:
				if rv == nil {
					return nil
				}
				return _ObjSerialize(*rv)
			default:
				panic(errors.New("suna.validator: not a json object"))
				return nil
			}
		},
		ParseValue: func(value interface{}) interface{} {
			m, e := ParseObject(value)
			if e != nil {
				panic(e)
			}
			return m
		},
		ParseLiteral: func(value ast.Value) interface{} {
			switch rv := value.(type) {
			case *ast.StringValue:
				m, e := ParseObject(rv.Value)
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

var GraphqlArray = graphql.NewScalar(
	graphql.ScalarConfig{
		Name: "Array",
		Serialize: func(value interface{}) interface{} {
			switch rv := value.(type) {
			case Array:
				return _ArySerialize(rv)
			case *Array:
				if rv == nil {
					return nil
				}
				return _ArySerialize(*rv)
			case []interface{}:
				return _ArySerialize(rv)
			case *[]interface{}:
				if rv == nil {
					return nil
				}
				return _ArySerialize(*rv)
			default:
				panic(errors.New("suna.validator: not a json array"))
				return nil
			}
		},
		ParseValue: func(value interface{}) interface{} {
			m, e := ParseArray(value)
			if e != nil {
				panic(e)
			}
			return m
		},
		ParseLiteral: func(value ast.Value) interface{} {
			switch rv := value.(type) {
			case *ast.StringValue:
				m, e := ParseArray(rv.Value)
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
