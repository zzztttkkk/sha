package sha

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
	"github.com/zzztttkkk/websocket"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

type _CustomFormTime time.Time

func (ft *_CustomFormTime) FromBytes(v []byte) bool {
	*ft = _CustomFormTime(time.Now())
	return true
}

type _CustomFormInt int64

func (fi *_CustomFormInt) FromBytes(v []byte) bool {
	i, e := strconv.ParseInt(utils.S(v), 10, 64)
	if e != nil {
		return false
	}
	*fi = _CustomFormInt(i)
	return true
}

type Sha5256Hash []byte

func (pwd *Sha5256Hash) FromBytes(v []byte) bool {
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
			_ = ctx.WriteString(strings.Repeat("Hello World!", 100))
		}),
	)

	mux.HTTP(
		"get",
		"/close",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			_ = ctx.WriteString("Hello World!")
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
			_ = ctx.WriteString("Hello World!")
		}),
	)

	mux.Websocket(
		"/ws",
		func(ctx context.Context, req *Request, conn *websocket.Conn) {
			for {
				_, d, e := conn.ReadMessage()
				if e != nil {
					break
				}
				fmt.Printf("recved from client: %s\n", d)

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
			_ = ctx.WriteString("hello world")
		}),
	)

	mux.HTTP(
		"get",
		"/chunked",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			f, e := os.Open("./server.go")
			if e != nil {
				ctx.SetError(e)
				return
			}
			defer f.Close()
			e = ctx.WriteStream(f)
			if e != nil {
				ctx.SetError(e)
			}
		}),
	)

	mux.HTTP(
		"get",
		"/compress_chunked",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			ctx.AutoCompress()

			f, e := os.Open("./server.go")
			if e != nil {
				ctx.SetError(e)
				return
			}
			defer f.Close()
			e = ctx.WriteStream(f)
			if e != nil {
				ctx.SetError(e)
			}
		}),
	)

	fmt.Println(mux)

	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGQUIT)
	defer cancelFunc()
	ListenAndServeWithContext(ctx, "127.0.0.1:8080", mux)
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

func TestPoolGC(t *testing.T) {
	ListenAndServe("", RequestHandlerFunc(func(ctx *RequestCtx) {
		fmt.Printf("%p\r\n", ctx)
		runtime.GC()
	}))
}
