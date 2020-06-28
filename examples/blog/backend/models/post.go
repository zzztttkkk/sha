package models

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/snow"
	"reflect"
	"time"

	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
	"github.com/zzztttkkk/snow/sqls"
)

type Post struct {
	sqls.Model
	Author   int64  `json:"author" ddl:"notnull"`
	Category int64  `json:"category" ddl:"notnull"`
	Title    string `json:"title" ddl:"notnull;L<50>"`
	Content  string `json:"content" ddl:"notnull;L<0>"`
}

type _PostOperatorT struct {
	sqls.Operator
}

var PostOperator = &_PostOperatorT{}

func init() {
	internal.LazyExecutor.Register(
		func(args snow.Kwargs) {
			PostOperator.Init(reflect.TypeOf(Post{}))
		},
	)
}

func (op *_PostOperatorT) Create(
	ctx context.Context, uid, categoryId int64,
	title, content []byte, tags ...int64,
) int64 {
	postId := op.XCreate(
		ctx,
		sqls.Dict{
			"created":  time.Now().Unix(),
			"author":   uid,
			"category": categoryId,
			"title":    title,
			"content":  content,
		},
	)
	op.AddTags(ctx, postId, tags...)
	return postId
}

func (op *_PostOperatorT) AddTags(ctx context.Context, postId int64, tagIds ...int64) {
	postTagsOperator.Create(ctx, postId, tagIds...)
}

func (op *_PostOperatorT) DelTags(ctx context.Context, postId int64, tagIds ...int64) {
	postTagsOperator.Delete(ctx, postId, tagIds...)
}

func (op *_PostOperatorT) Update(ctx context.Context, uid, postId int64, dict sqls.Dict) bool {
	placeholder, values := dict.ForUpdate()
	values = append(values, uid)
	values = append(values, postId)

	return op.XUpdate(
		ctx,
		fmt.Sprintf(`update post set(%s) where author_id=? and id=?`, placeholder),
		values...,
	) > 0
}

func (op *_PostOperatorT) Delete(ctx context.Context, postId int64) bool {
	return op.XUpdate(
		ctx,
		`update post set(deleted=?) where id=? and deleted=0`,
		time.Now().Unix(), postId,
	) > 0
}
