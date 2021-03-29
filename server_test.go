package sha

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
	"github.com/zzztttkkk/websocket"
	"net/http"
	_ "net/http/pprof"
	"regexp"
	"runtime"
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
	mux := NewMux(nil)
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
			ctx.Response.Header().Set(HeaderConnection, []byte("close"))
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

	mux.HTTPWithOptions(
		&HandlerOptions{Document: validator.NewDocument(Form{}, validator.Undefined)},
		"post",
		"/form",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.Response.Header().SetContentType(MIMEText)
			var form Form
			ctx.MustValidateForm(&form)
			fmt.Printf("%s\n", form.Password)
			for _, file := range ctx.Request.Files() {
				fmt.Println(file.Header)
				fmt.Println(file.FileName, file.Name)
				fmt.Println(string(file.Data()))
			}
			_, _ = ctx.WriteString("Hello World!")
		}),
	)

	mux.Websocket(
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
		nil,
	)

	mux.HTTP(
		"get",
		"/hello",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("hello world")
		}),
	)

	fmt.Println(mux)
	ListenAndServe("127.0.0.1:8080", mux)
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

func TestMergo(t *testing.T) {
	type A struct {
		Z string `json:"z"`
		X int64  `json:"-"`
	}

	var a1 = A{Z: "a1", X: 45}
	var a2 A
	a2.X = 56

	err := mergo.Merge(&a2, &a1)
	if err != nil {
		panic(err)
	}
	fmt.Println(&a1, &a2)
}

func TestPoolGC(t *testing.T) {
	ListenAndServe("", RequestHandlerFunc(func(ctx *RequestCtx) {
		fmt.Printf("%p\r\n", ctx)
		runtime.GC()
	}))
}
