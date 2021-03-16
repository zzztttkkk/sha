package internal

import (
	"context"
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stderr, "sha.rbac ", log.LstdFlags)
}

type _RootCtxKeyT int

const RootCtxKey = _RootCtxKeyT(0)

type RootCtx struct {
	context.Context
	Uid  int64
	Info interface{}
}

func NewRootContext(id int64, info interface{}) context.Context {
	v := &RootCtx{Uid: id, Info: info}
	v.Context = context.WithValue(context.Background(), RootCtxKey, v)
	return v
}
