package suna

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/validator"
	"github.com/zzztttkkk/websocket"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestServer_Run(t *testing.T) {
	mux := NewMux("", nil)
	mux.AutoOptions = true
	mux.AutoSlashRedirect = true
	mux.REST(
		"get",
		"/compress",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			_, _ = ctx.WriteString(strings.Repeat("Hello World!", 100))
		}),
	)

	mux.REST(
		"get",
		"/close",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("Hello World!")
			ctx.Response.Header.SetStr(HeaderConnection, "close")
		}),
	)

	validator.RegisterRegexp("joineduints", regexp.MustCompile(`\d+(,\d+)*`))

	type Form struct {
		String string `validator:",L=3"`
		Bytes  []byte `validator:",R=joineduints"`
		Int    int64  `validator:",V=40-60"`
		Bool   bool   `validator:",optional"`
	}

	mux.REST(
		"get",
		"/form",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form Form
			ctx.MustValidate(&form)
			fmt.Printf("%+v\n", form)
			_, _ = ctx.WriteString("Hello World!")
		}),
	)

	mux.WebSocket(
		"/ws",
		func(ctx context.Context, req *Request, conn *websocket.Conn) {
			for {
				_, d, e := conn.ReadMessage()
				if e != nil {
					break
				}
				fmt.Println(string(d))

				if e = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", time.Now()))); e != nil {
					break
				}
			}
		},
	)

	mux.Print(false, false)
	server := Default(mux)
	server.ListenAndServe()
}
