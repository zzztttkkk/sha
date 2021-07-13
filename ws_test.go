package sha

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/zzztttkkk/websocket"
)

var mux *Mux

func init() {
	mux = NewMux(nil)
	mux.HTTP(
		MethodGet,
		"/",
		RequestCtxHandlerFunc(func(ctx *RequestCtx) {
			res := ctx.Response
			res.Header().SetContentType(MIMEHtml)
			f, _ := os.Open("./ws.html")
			p, _ := ioutil.ReadAll(f)
			_, _ = ctx.Write(p)
		}),
	)
	mux.Websocket(
		"/ws",
		func(ctx context.Context, req *Request, conn *websocket.Conn) {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						i, p, e := conn.ReadMessage()
						if e != nil {
							log.Println(e)
							return
						}
						fmt.Println(i, string(p))
					}
				}
			}()

			for {
				conn.Subprotocol()
				time.Sleep(time.Second)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(time.Now().String()))
			}
		},
		nil,
	)
}

func TestWebSocketProtocol_Serve(t *testing.T) {
	s := Default()
	s.Handler = mux
	s.ListenAndServe()
}

func TestWebSocketProtocol_ServeTLS(t *testing.T) {
	conf := ServerOptions{}
	conf.TLS.Key = "./tls/sha.local-key.pem"
	conf.TLS.Cert = "./tls/sha.local.pem"
	s := New(context.Background(), nil, &conf)
	s.Handler = mux
	s.ListenAndServe()
}
