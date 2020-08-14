package builder

import (
	"fmt"
	"github.com/zzztttkkk/sqrl"
	"testing"
)

func Test_Gen(t *testing.T) {
	isPostgres = true

	cond := AndConditions()
	fmt.Println(cond.ToSql())

	c1 := AndConditions()
	c1.Eq(true, "id", 34)
	c1.Like(true, "name", "aaa%")
	fmt.Println(c1.ToSql())

	c2 := OrConditions()
	c2.Lt(true, "created", 111)
	c2.Gt(true, "created", 333)
	fmt.Println(c2.ToSql())

	fmt.Println(NewSelect("*").From("users").Where(And(c1, c2)).ToSql())

	c3 := AndConditions()
	c3.Eq(true, "x", sqrl.Raw("now()"))
	fmt.Println(c3.ToSql())

	ub := NewUpdate("ss").
		Set("b", sqrl.Expr("json_array_append(b, '$.history', ?)", 3434)).
		Set("a", sqrl.Raw("now()+1"))

	fmt.Println(ub.ToSql())
}
