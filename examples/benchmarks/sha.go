package main

import "github.com/zzztttkkk/sha"

type Sha struct{}

func (_ Sha) Name() string {
	return "sha"
}

func (_ Sha) HelloWorld(address string) {
	sha.ListenAndServe(address, sha.RequestCtxHandlerFunc(func(ctx *sha.RequestCtx) {
		_ = ctx.WriteString("HelloWorld!")
	}))
}

var _ Engine = Sha{}

func init() { register(Sha{}) }
