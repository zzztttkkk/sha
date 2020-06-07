package sqls

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTableDefinition(t *testing.T) {
	type User struct {
		Id    int64  `ddl:":primary;incr;"`
		Name  string `ddl:":L<30>;unique"`
		Alias string `ddl:":L<30>;D<'%3B'>"`
	}
	fmt.Println(TableDefinition(reflect.TypeOf(User{})))
}
