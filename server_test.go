package sha

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"net/http"
	_ "net/http/pprof"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/zzztttkkk/sha/validator"
	"github.com/zzztttkkk/websocket"
)

type _CustomFormTime time.Time

func (ft *_CustomFormTime) FormValue(v []byte) bool {
	*ft = _CustomFormTime(time.Now())
	return true
}

type _CustomFormInt int64

func (fi *_CustomFormInt) FormValue(v []byte) bool {
	i, e := strconv.ParseInt(utils.S(v), 10, 64)
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
	mux.HTTP(
		"get",
		"/compress",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()
			_, _ = ctx.WriteString(strings.Repeat("Hello World!", 100))
		}),
	)

	mux.HTTP(
		"get",
		"/close",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("Hello World!")
			ctx.Response.Header.Set(HeaderConnection, []byte("close"))
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

	mux.HTTPWithForm(
		"post",
		"/form",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.Response.Header.SetContentType(MIMEText)
			var form Form
			ctx.MustValidate(&form)
			fmt.Printf("%s\n", form.Password)
			for _, file := range ctx.Request.Files() {
				fmt.Println(file.Header)
				fmt.Println(file.FileName, file.Name)
				fmt.Println(string(file.Data()))
			}
			_, _ = ctx.WriteString("Hello World!")
		}),
		Form{},
	)

	mux.WebSocket(
		"/ws",
		func(ctx context.Context, req *Request, conn *websocket.Conn, _ string) {
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

	mux.HTTP(
		"get",
		"/hello",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("hello world")
		}),
	)

	mux.Print()
	server := Default(mux)
	server.ListenAndServe()
}

func TestServer_RunSimple(t *testing.T) {
	server := Default(RequestHandlerFunc(func(ctx *RequestCtx) {
		_, _ = ctx.WriteString("hello world")
	}))
	server.ListenAndServe()
}

func TestServer_RunPrintRequest(t *testing.T) {
	server := Default(RequestHandlerFunc(func(ctx *RequestCtx) {
		req := &ctx.Request

		fmt.Printf(
			"%s %s\n%s\n%s\n%s\n",
			req.Method,
			req.Path,
			req.Query(),
			&req.Header,
			req.BodyForm(),
		)

		_, _ = ctx.WriteString("hello world")
	}))
	server.ListenAndServe()
}

func TestServer_RunTimeout(t *testing.T) {
	server := Default(RequestHandlerFunc(func(ctx *RequestCtx) { _, _ = ctx.WriteString("hello world") }))
	var fn func()
	server.BaseCtx, fn = context.WithTimeout(context.Background(), time.Second)
	defer fn()
	server.ListenAndServe()
}

func TestStdHttp(t *testing.T) {
	s := &http.Server{
		IdleTimeout: time.Second * 30,
		Addr:        "127.0.0.1:8080",
	}
	s.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello world"))
	})
	_ = s.ListenAndServe()
}
