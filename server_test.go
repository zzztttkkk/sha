package sha

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/validator"
	"github.com/zzztttkkk/websocket"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

type _CustomFormTime time.Time

func (ft *_CustomFormTime) FormValue(v []byte) bool {
	*ft = _CustomFormTime(time.Now())
	return true
}

type _CustomFormInt int64

func (fi *_CustomFormInt) FormValue(v []byte) bool {
	i, e := strconv.ParseInt(internal.S(v), 10, 64)
	if e != nil {
		return false
	}
	*fi = _CustomFormInt(i)
	return true
}

type Sha5256Hash []byte

func (pwd *Sha5256Hash) FormValue(v []byte) bool {
	n := sha512.New512_256()
	n.Write(v)
	dist := make([]byte, 64)
	hex.Encode(dist, n.Sum(nil))
	*pwd = dist
	return true
}

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

	validator.RegisterRegexp("joineduints", regexp.MustCompile(`^\d+(,\d+)*$`))
	validator.RegisterRegexp("password", regexp.MustCompile(`^\w{6,}$`))

	type Form struct {
		FormTime _CustomFormTime `validator:"ft"`
		FormInt  _CustomFormInt  `validator:"fi"`
		String   string          `validator:",L=3"`
		Bytes    []byte          `validator:",R=joineduints"`
		Int      int64           `validator:",V=40-60"`
		Bool     bool            `validator:",optional"`
		Password Sha5256Hash     `validator:"pwd,R=password"`
	}

	mux.RESTWithForm(
		"get",
		"/form",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form Form
			ctx.MustValidate(&form)
			fmt.Printf("%+v\n", form)
			_, _ = ctx.WriteString("Hello World!")
		}),
		Form{},
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
