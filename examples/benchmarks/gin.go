package main

import "github.com/gin-gonic/gin"

type Gin struct{}

func (_ Gin) Name() string {
	return "gin"
}

func (_ Gin) HelloWorld(address string) {
	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	server.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "HelloWorld!")
	})
	_ = server.Run(address)
}

var _ Engine = Gin{}

func init() {
	register(Gin{})
}
