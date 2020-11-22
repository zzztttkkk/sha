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

func init() {
	var err error
	wdb, err = sql.Open("mysql", "root:123456@/suna?parseTime=true")
	if err != nil {
		panic(err)
	}
}

func Test_Row(t *testing.T) {
	var err error

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
}

func Test_ExcScanner_Rows(t *testing.T) {
	var err error
	ctx := context.Background()
	es := ExcScanner(ctx)

	// primitive
	var ids []int
	err = es.Rows(ctx, StrSql("select id from user"), &ids)
	fmt.Println(ids, err)

	// primitive ptr
	var idps []*int
	err = es.Rows(ctx, StrSql("select id from user"), &idps)
	fmt.Println(idps, err)
	for _, p := range idps {
		fmt.Println(*p)
	}

	// time.Time
	var cas []time.Time
	err = es.Rows(ctx, StrSql("select created_at from user"), &cas)
	fmt.Println(cas, err)

	// *time.Time
	var caps []*time.Time
	err = es.Rows(ctx, StrSql("select created_at from user"), &caps)
	fmt.Println(caps, err)
	for _, p := range caps {
		fmt.Println(*p)
	}

	// unscannable struct
	var as []A
	err = es.Rows(ctx, StrSql("select * from user"), &as)
	fmt.Println(as, err)

	// unscannable struct ptr
	var aps []*A
	err = es.Rows(ctx, StrSql("select * from user"), &aps)
	fmt.Println(aps, err)
	for _, p := range aps {
		fmt.Println(*p)
	}

	// scannable struct
	var cans []sql.NullTime
	err = es.Rows(ctx, StrSql("select created_at from user"), &cans)
	fmt.Println(cans, err)

	// scannable struct ptr
	var canps []*sql.NullTime
	err = es.Rows(ctx, StrSql("select created_at from user"), &canps)
	fmt.Println(canps, err)
	for _, p := range canps {
		fmt.Println(*p)
	}
}
