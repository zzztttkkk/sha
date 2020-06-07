package models

import (
	"fmt"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
)

type model struct {
	Id      int64 `json:"id" ddl:"incr;primary"`
	Status  int   `json:"status" ddl:"D<0>"`
	Created int64 `json:"created"`
	Deleted int64 `json:"deleted" ddl:"D<0>"`
}

type User struct {
	model
	Name     string `json:"name" ddl:"notnull;unique;L<30>"`
	Alias    string `json:"alias" ddl:"L<30>;D<''>"`
	Password []byte `json:"password" ddl:"notnull;L<64>"`
	Bio      string `json:"bio" ddl:"L<120>;D<''>"`
	Avatar   string `json:"avatar" ddl:"L<120>;D<''>"`
}

type Category struct {
	model
	Name  string `json:"name" ddl:"notnull;unique;L<30>"`
	Descp string `json:"descp" ddl:"L<0>;notnull"`
}

type Tag struct {
	model
	Name  string `json:"name" ddl:"notnull;unique;L<30>"`
	Descp string `json:"descp" ddl:"L<0>;notnull"`
}

type Post struct {
	model
	Category int64  `json:"category" ddl:"notnull"`
	Title    string `json:"title" ddl:"notnull;L<50>"`
	Content  string `json:"content" ddl:"notnull;L<0>"`
}

type PostAndTag struct {
	TagId  int64 `ddl:"primary"`
	PostId int64 `ddl:"primary"`
}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(User{})))
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(Category{})))
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(Tag{})))
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(Post{})))
			sqls.Master().MustExec(sqls.TableDefinition(reflect.TypeOf(PostAndTag{})))
			fmt.Println(232323)
		},
	)
	fmt.Println(234234)
}
