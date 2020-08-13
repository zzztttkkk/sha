package builder

import (
	"fmt"
	"github.com/zzztttkkk/sqrl"
	"github.com/zzztttkkk/suna/sqls"
	"testing"
)

func Test_Gen(t *testing.T) {
	sqls.isPostgres = true

	cond := Builder.And()
	fmt.Println(cond.ToSql())

	c1 := Builder.And()
	c1.Eq(true, "id", 34)
	c1.Like(true, "name", "aaa%")
	fmt.Println(c1.ToSql())

	c2 := Builder.Or()
	c2.Lt(true, "created", 111)
	c2.Gt(true, "created", 333)
	fmt.Println(c2.ToSql())

	fmt.Println(Builder.NewSelect("*").From("users").Where(And(c1, c2)).ToSql())

	c3 := Builder.And()
	c3.Eq(true, "x", sqrl.Raw("now()"))
	fmt.Println(c3.ToSql())
}
