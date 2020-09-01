package builder

import (
	"fmt"
	"github.com/zzztttkkk/sqlr"
	"testing"
)

func Test_Gen(t *testing.T) {
	isPostgres = true

	cond := AND()
	fmt.Println(cond.ToSql())

	c1 := AND()
	c1.Eq(true, "id", 34)
	c1.Like(true, "name", "aaa%")
	fmt.Println(c1.ToSql())

	c2 := OR()
	c2.Lt(true, "created", 111)
	c2.Gt(true, "created", 333)
	fmt.Println(c2.ToSql())

	fmt.Println(NewSelect("*").From("users").Where(sqlr.And{c1, c2}).ToSql())

	c3 := AND()
	c3.Eq(true, "x", sqlr.Raw("now()"))
	fmt.Println(c3.ToSql())

	ub := NewUpdate("ss").
		Set("b", sqlr.Expr("json_array_append(b, '$.history', ?)", 3434)).
		Set("a", sqlr.Raw("now()+1"))

	fmt.Println(ub.ToSql())
}
