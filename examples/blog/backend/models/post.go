package models

import (
	"context"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
	"time"
)

type Post struct {
	sqls.Model
	AuthorId int64  `json:"author_id" ddl:"notnull"`
	Category int64  `json:"category" ddl:"notnull"`
	Title    string `json:"title" ddl:"notnull;L<50>"`
	Content  string `json:"content" ddl:"notnull;L<0>"`
}

type _PostOperatorT struct {
	sqls.Operator
}

var PostOperator = &_PostOperatorT{}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			PostOperator.Init(reflect.TypeOf(Post{}))
			PostOperator.SqlsTableCreate()
		},
	)
}

func (op *_PostOperatorT) Create(ctx context.Context, uid, categoryId int64, title, content []byte) int64 {
	return op.SqlxCreate(
		ctx,
		`insert into post (created,author_id,category,title,content) values(?,?,?,?,?)`,
		time.Now().Unix(), uid, categoryId,
		title, content,
	)
}

func (op *_PostOperatorT) AddTag(ctx context.Context, postId, tagId int64) {

}
