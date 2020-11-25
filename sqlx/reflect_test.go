package sqlx

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"net/url"
	"testing"
	"time"
)

type TestModel struct {
	Id        int64      `db:"id,g=login" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
	Name      Bytes      `db:"name,g=login" json:"name"`
	Password  Bytes      `db:"password,g=login" json:"password"`
}

func (TestModel) TableColumns(db *sqlx.DB) []string {
	return []string{
		"id bigint primary key auto_increment",
		"created_at datetime default now()",
		"deleted_at datetime",
		"name char(20) unique not null",
		"password char(128)",
	}
}

var TestModelOperator *Operator

func init() {
	OpenWriteableDB("mysql", "root:123456@/suna?parseTime=true&loc="+url.QueryEscape("Asia/Shanghai"))
	EnableLogging()

	TestModelOperator = NewOperator(TestModel{})
	TestModelOperator.CreateTable(true)
}

func Test_XWrapper_Exe(t *testing.T) {
	ctx := context.Background()

	var m TestModel
	_ = TestModelOperator.One(ctx, "*", "WHERE id=1", nil, &m)
	fmt.Println(m, string(m.Name))
	j, e := json.Marshal(&m)
	fmt.Println(e, string(j), 111)

	var name, password []byte
	_ = Exe(ctx).Row(ctx, "select name,password from test_model where id=1 and deleted_at is null", nil, &name, &password)
	fmt.Println(string(name), string(password))

	fmt.Println(TestModelOperator.GroupKeys("login"))
}

func TestOperator_Insert(t *testing.T) {
	ctx := context.Background()
	tctx, committer := Tx(ctx)
	defer committer()

	r := TestModelOperator.Insert(tctx, M{"name": "pou", "created_at": Raw("DATE_ADD(NOW(), INTERVAL 31 DAY)")})
	fmt.Println(r.LastInsertId())
}
