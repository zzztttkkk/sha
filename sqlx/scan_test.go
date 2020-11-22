package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

type A struct {
	Id int
	CA time.Time `sqlx:"created_at"`
}

func Test_Scan(t *testing.T) {
	var err error
	wdb, err = sql.Open("mysql", "root:123456@/suna?parseTime=true")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	es := ExcScanner(ctx)

	var a A
	err = es.Row(ctx, StrSql("select id,created_at from user limit 1"), &a)
	fmt.Println(a, err)

	var id int
	var ca time.Time
	err = es.Row(ctx, StrSql("select id,created_at from user limit 1"), &id, &ca)
	fmt.Println(id, ca, err)
	ca = time.Time{}
	err = es.Row(ctx, StrSql("select created_at from user limit 1"), &ca)
	fmt.Println(ca, err)
	var nca sql.NullTime
	err = es.Row(ctx, StrSql("select created_at from user limit 1"), &nca)
	fmt.Println(nca, err)

	var as []A
	err = es.Rows(ctx, StrSql("select * from user"), &as)
	fmt.Println(as, err)

	var aps []*A
	err = es.Rows(ctx, StrSql("select * from user"), &aps)
	fmt.Println(aps, err)
	for _, p := range aps {
		fmt.Println(*p)
	}

	var ids []int
	err = es.Rows(ctx, StrSql("select id from user"), &ids)
	fmt.Println(ids, err)

	var idps []*int
	err = es.Rows(ctx, StrSql("select id from user"), &idps)
	fmt.Println(idps, err)
	for _, p := range idps {
		fmt.Println(*p)
	}

	var cas []time.Time
	err = es.Rows(ctx, StrSql("select created_at from user"), &cas)
	fmt.Println(cas, err)

	var caps []*time.Time
	err = es.Rows(ctx, StrSql("select created_at from user"), &caps)
	fmt.Println(caps, err)
	for _, p := range caps {
		fmt.Println(*p)
	}
}
