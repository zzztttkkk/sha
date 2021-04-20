package sqlx

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	var buf strings.Builder
	writeCond("   WHERE 1=1", &buf)

	fmt.Println(buf.String())
}

type TestTxModel struct {
	Num int64 `db:"num"`
}

func (t TestTxModel) TableName() string {
	return "test_tx"
}

func (t TestTxModel) TableColumns(db *sqlx.DB) []string {
	return []string{"num bigint primary key not null"}
}

var _ Modeler = TestTxModel{}

func TestTxWithOptions(t *testing.T) {
	var op = NewOperator(TestTxModel{})
	op.CreateTable()

	ctx, tx := Tx(context.Background())
	defer tx.Commit(ctx)

	op.Delete(ctx, "1=1", nil)

	ctx, tx = TxWithOptions(ctx, &TxOptions{SavePointName: "t1"}) // savepoint sha_sqlx_sub_tx_begin
	op.Insert(ctx, Data{"num": 55})
	tx.Commit(ctx) // savepoint t1

	ctx, tx = TxWithOptions(ctx, &TxOptions{SavePointName: "t2"}) // savepoint sha_sqlx_sub_tx_begin
	op.Insert(ctx, Data{"num": 55})
	tx.Commit(ctx) // savepoint t2
}
