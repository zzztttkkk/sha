package grafql

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/savsgio/gotils"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/ctxs"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"strings"
	"testing"
)

type B struct {
	Id   int64
	Name string
}

func (b *B) ResolveOutName(p graphql.ResolveParams) (interface{}, error) {
	rbac.IsGranted(
		p.Context,
		auth.MustGetUser(ctxs.MustUnwrap(p.Context)),
		rbac.PolicyAny,
		"b.name.read", "b.name.*", "b.*.read", "b.*.*",
	)
	return b.Name, nil
}

type C struct {
	CId   int64
	CName string
}

func _ParseC(data string) interface{} {
	s := strings.Split(data, ":")
	if len(s) != 2 {
		return nil
	}
	return &C{CId: utils.S2I64(s[0]), CName: s[1]}
}

func (C) GraphqlScalar() *graphql.Scalar {
	return graphql.NewScalar(
		graphql.ScalarConfig{
			Name:        "C",
			Description: "",
			Serialize: func(value interface{}) interface{} {
				switch rv := value.(type) {
				case C:
					return fmt.Sprintf("%d:%s", rv.CId, rv.CName)
				case *C:
					if rv == nil {
						return nil
					}
					return fmt.Sprintf("%d:%s", rv.CId, rv.CName)
				default:
					return nil
				}
			},
			ParseValue: func(value interface{}) interface{} {
				var data string

				switch rv := value.(type) {
				case string:
					data = rv
				case *string:
					if rv == nil {
						return nil
					}
					data = *rv
				case []byte:
					data = gotils.B2S(rv)
				case *[]byte:
					if rv == nil {
						return nil
					}
					data = gotils.B2S(*rv)
				default:
					return nil
				}

				return _ParseC(data)
			},
			ParseLiteral: func(value ast.Value) interface{} {
				switch rv := value.(type) {
				case *ast.StringValue:
					return _ParseC(rv.Value)
				default:
					return nil
				}
			},
		},
	)
}

type D struct {
	B
	C
}

func TestNewObject(t *testing.T) {
	v := NewOutObject(reflect.ValueOf(D{}))
	fmt.Println(v)
}
