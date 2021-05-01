package sqlx

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/zzztttkkk/sha/jsonx"
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
	//OpenWriteableDB("mysql", "root:123456@/sha?parseTime=true&loc="+url.QueryEscape("Asia/Shanghai"))
	OpenWriteableDB("postgres", "postgres://postgres:123456@127.0.0.1:5432/sha?sslmode=disable")
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
	ctx, tx := Tx(ctx)
	defer tx.AutoCommit(ctx)

	r, _ := TestModelOperator.Insert(ctx, Data{"name": "pou", "created_at": Raw("DATE_ADD(NOW(), INTERVAL 31 DAY)")})
	fmt.Println(r)
}

type TestModel2 struct {
	ID        int64      `db:"id,g=login" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	M1ID      int64      `db:"m1_id" json:"m1_id"`
	M1        *TestModel `db:"-" json:"m"`
}

func (t TestModel2) TableName() string {
	return "test_model2"
}

func (t TestModel2) TableColumns(db *sqlx.DB) []string {
	return []string{
		"id serial primary key",
		"created_at timestamp default now()",
		"m1_id bigint not null",
	}
}

var TestModel2Operator *Operator

func init() {
	TestModel2Operator = NewOperator(TestModel2{})
	TestModel2Operator.CreateTable()
}
func TestJoin(t *testing.T) {
	ctx := context.Background()
	var m TestModel2
	var m1 TestModel
	m.M1 = &m1
	type Arg struct {
		AID int64 `db:"aid"`
	}
	e := Exe(ctx).JoinGet(
		ctx, "select * from test_model, test_model2 where test_model2.m1_id = test_model.id and test_model2.id=:aid", Arg{AID: 1},
		m.M1, &m,
	)
	if e != nil {
		panic(e)
	}
	if m.M1.Password == "" {
		t.Fail()
	}

	var pL []*TestModel2
	e = Exe(ctx).JoinSelect(
		ctx,
		"select * from test_model, test_model2 where test_model2.m1_id = test_model.id", nil, &pL,
		func(i interface{}) {
			v := i.(*TestModel2)
			v.M1 = &TestModel{}
		},
		func(i interface{}, i2 int) interface{} {
			v := i.(*TestModel2)
			switch i2 {
			case 0:
				return v.M1
			case 1:
				return v
			default:
				panic("")
			}
		},
	)
	if e != nil {
		panic(e)
	}
	fmt.Println(pL[0])

	var vL = make([]TestModel2, 0, 10)
	e = Exe(ctx).JoinSelect(
		ctx,
		"select * from test_model, test_model2 where test_model2.m1_id = test_model.id", nil, &vL,
		func(i interface{}) {
			v := i.(*TestModel2)
			v.M1 = &TestModel{}
		},
		func(i interface{}, i2 int) interface{} {
			v := i.(*TestModel2)
			switch i2 {
			case 0:
				return v.M1
			case 1:
				return v
			default:
				panic("")
			}
		},
	)
	if e != nil {
		panic(e)
	}
	fmt.Println(vL[0])
}
