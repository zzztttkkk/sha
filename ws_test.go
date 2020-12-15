package sha

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/websocket"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

var mux *Mux

func init() {
	mux = NewMux("", nil)
	mux.HTTP(
		MethodGet,
		"/",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			res := ctx.Response
			res.Header.SetContentType(MIMEHtml)
			f, _ := os.Open("./ws.html")
			p, _ := ioutil.ReadAll(f)
			_, _ = ctx.Write(p)
		}),
	)
	mux.WebSocket(
		"/ws",
		func(ctx context.Context, req *Request, conn *websocket.Conn, _ string) {
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
				time.Sleep(time.Second)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", time.Now())))
			}
		},
	)
}

func TestWebSocketProtocol_Serve(t *testing.T) {
	s := Default(mux)
	s.ListenAndServe()
}

func TestWebSocketProtocol_ServeTLS(t *testing.T) {
	s := Default(mux)
	s.ListenAndServeTLS("./tls/ztk.local+3.pem", "./tls/ztk.local+3-key.pem")
}
