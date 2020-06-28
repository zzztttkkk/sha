package sqls

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
)

func TestTableDefinition(t *testing.T) {
	type User struct {
		Id    int64  `ddl:":primary;incr;"`
		Name  string `ddl:":L<30>;unique"`
		Alias string `ddl:":L<30>;D<'%3B'>"`
		Descp sql.NullString
	}
	fmt.Println(TableDefinition(reflect.TypeOf(User{})))
}
