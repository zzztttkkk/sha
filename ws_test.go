package sha

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/websocket"
	"log"
	"testing"
	"time"
)

func TestWebSocketProtocol_Serve(t *testing.T) {
	s := Default(nil)
	mux := NewMux("", nil)
	mux.WebSocket(
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
				time.Sleep(time.Second)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", time.Now())))
			}
		},
	)

	s.Handler = mux

	s.ListenAndServe()
}
