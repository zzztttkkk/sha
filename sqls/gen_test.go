package sqls

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_Gen(t *testing.T) {
	Gen("blog:123456@tcp(127.0.0.1:3306)/blog")
}
