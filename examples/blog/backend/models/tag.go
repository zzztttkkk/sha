package models

import (
	"context"
	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
	"reflect"
	"time"
)

type Tag struct {
	sqls.Enum
	Descp string `json:"descp" ddl:"L<0>;notnull"`
}

type _PostTags struct {
	TagId   int64 `ddl:"primary"`
	PostId  int64 `ddl:"primary"`
	Created int64 `ddl:"notnull"`
}

func (p *_PostTags) TableName() string {
	return "post_tags"
}

type _TagOperatorT struct {
	sqls.EnumOperator
}

var TagOperator = &_TagOperatorT{}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			TagOperator.Init(
				reflect.TypeOf(Tag{}),
				func() sqls.Enumer { return &Tag{} },
			)
		},
	)
}

type _PostTagsOperatorT struct {
	sqls.Operator
}

var postTagsOperator = &_PostTagsOperatorT{}

func init() {
	internal.LazyE.Register(
		func(args ...interface{}) {
			postTagsOperator.Init(reflect.TypeOf(_PostTags{}))
		},
	)
}

func (op *_PostTagsOperatorT) Create(ctx context.Context, postId int64, tagsId ...int64) {
	stmt := op.SqlxStmt(ctx, `insert into post_tags (tagid,postid,created) values(?,?,?)`)
	defer stmt.Close()

	now := time.Now().Unix()
	for _, tid := range tagsId {
		_, err := stmt.ExecContext(ctx, tid, postId, now)
		if err != nil {
			panic(err)
		}
	}
}

func (op *_PostTagsOperatorT) Delete(ctx context.Context, postId int64, tagsId ...int64) {
	stmt := op.SqlxStmt(ctx, `delete from post_tags where postid=? and tagid=?`)
	defer stmt.Close()

	for _, tid := range tagsId {
		_, err := stmt.ExecContext(ctx, tid, postId)
		if err != nil {
			panic(err)
		}
	}
}
