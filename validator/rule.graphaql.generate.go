package validator

import (
	"errors"
	"github.com/graphql-go/graphql"
	"github.com/zzztttkkk/suna/jsonx"
)

func (rule *_Rule) toGraphqlArgument() *graphql.ArgumentConfig {
	cfg := &graphql.ArgumentConfig{Description: rule.info}

	switch rule.t {
	case _Bool:
		cfg.Type = graphql.Boolean
		if len(rule.defaultV) > 0 {
			v, _ := rule.toBool(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _Int64:
		cfg.Type = graphql.Int
		if len(rule.defaultV) > 0 {
			v, _ := rule.toInt(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _Uint64:
		cfg.Type = graphql.Int
		if len(rule.defaultV) > 0 {
			v, _ := rule.toUint(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _Float64:
		cfg.Type = graphql.Float
		if len(rule.defaultV) > 0 {
			v, _ := rule.toFloat(rule.defaultV)
			cfg.DefaultValue = v
		}
	case _String:
		cfg.Type = graphql.String
		if len(rule.defaultV) > 0 {
			v, _ := rule.toBytes(rule.defaultV)
			cfg.DefaultValue = string(v)
		}
	case _Bytes, _BytesSlice:
		panic(errors.New("suna.validator[graphql]: use `string` instead of `[]byte`"))
	case _JsonArray:
		cfg.Type = jsonx.GraphqlArray
	case _JsonObject:
		cfg.Type = jsonx.GraphqlObject
	case _BoolSlice:
		cfg.Type = graphql.NewList(graphql.Boolean)
	case _IntSlice, _UintSlice:
		cfg.Type = graphql.NewList(graphql.Int)
	case _FloatSlice:
		cfg.Type = graphql.NewList(graphql.Float)
	case _StringSlice:
		cfg.Type = graphql.NewList(graphql.String)
	}

	if rule.required {
		cfg.Type = graphql.NewNonNull(cfg.Type)
	}

	return cfg
}

func (rs *Rules) ArgumentMap() graphql.FieldConfigArgument {
	m := graphql.FieldConfigArgument{}
	for _, r := range rs.lst {
		m[r.form] = r.toGraphqlArgument()
	}
	return m
}
