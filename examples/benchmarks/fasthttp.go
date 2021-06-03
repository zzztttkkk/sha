package main

import "github.com/valyala/fasthttp"

type Fasthttp struct{}

func (Fasthttp) Name() string { return "fasthttp" }

func (Fasthttp) HelloWorld(address string) {
	_ = fasthttp.ListenAndServe(address, func(ctx *fasthttp.RequestCtx) { _, _ = ctx.WriteString("HelloWorld!") })
}

var _ Engine = Fasthttp{}

func init() {
	register(Fasthttp{})
}
