package sqlx

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/zzztttkkk/sha/jsonx"
	"net/url"
	"testing"
	"time"
)

type TestModel struct {
	ID        int64      `db:"id,g=login" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
	Name      string     `db:"name,g=login" json:"name"`
	Password  string     `db:"password,g=login" json:"password"`
}

func (TestModel) TableName() string { return "test_model" }

func (TestModel) TableColumns(db *sqlx.DB) []string {
	switch db.DriverName() {
	case "mysql":
		return []string{
			"id bigint primary key auto_increment",
			"created_at datetime default now()",
			"deleted_at datetime",
			"name char(20) unique not null",
			"password char(128)",
		}
	case "postgres":
		return []string{
			"id serial primary key",
			"created_at timestamp default now()",
			"deleted_at timestamp",
			"name char(20) unique not null",
			"password char(128)",
		}
	default:
		return nil
	}
}

var TestModelOperator *Operator

func init() {
	OpenWriteableDB("mysql", "root:123456@/sha?parseTime=true&loc="+url.QueryEscape("Asia/Shanghai"))
	//OpenWriteableDB("postgres", "postgres://postgres:123456@127.0.0.1:5555/sha?sslmode=disable")
	EnableLogging()

	TestModelOperator = NewOperator(TestModel{})
	TestModelOperator.CreateTable()
}

func Test_XWrapper_Exe(t *testing.T) {
	ctx := context.Background()

	var m TestModel
	type Arg struct {
		UID int64 `db:"user_id"`
	}

	_ = TestModelOperator.FetchOne(ctx, "*", "WHERE id=:user_id", Arg{UID: 12}, &m)
	fmt.Println(m, m.Name)
	j, e := jsonx.Marshal(&m)
	fmt.Println(e, string(j), 111)

	var name, password []byte
	_ = Exe(ctx).ScanRow(ctx, "select `name`,`password` from test_model where id=1 and deleted_at is null", nil, &name, &password)
	fmt.Println(string(name), string(password))

	fmt.Println(TestModelOperator.GroupColumns("login"))
}

func TestOperator_Insert(t *testing.T) {
	ctx := context.Background()
	tctx, tw := Tx(ctx)
	defer tw.Commit(tctx)

	r, _ := TestModelOperator.Insert(tctx, Data{"name": "pou", "created_at": Raw("DATE_ADD(NOW(), INTERVAL 31 DAY)")})
	fmt.Println(r)
}
