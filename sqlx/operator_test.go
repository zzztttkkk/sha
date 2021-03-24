package sqlx

import (
	"context"
	"fmt"
	"testing"
)

func Test_Marge(t *testing.T) {
	fmt.Println(mergeMap(context.Background(), Data{}, Data{"a": 45}))
	fmt.Println(mergeMap(context.Background(), Data{}, map[string]interface{}{"a": 45}))
	type Arg struct {
		Int    int64  `db:"number"`
		String string `db:"string"`
	}
	fmt.Println(mergeMap(context.Background(), Data{"m": 567}, Arg{Int: 1234, String: "Asd"}))
}
