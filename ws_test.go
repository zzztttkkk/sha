package suna

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
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				default:
					i, p, e := conn.ReadMessage()
					if e != nil {
						log.Println(e)
						break loop
					}
					fmt.Println(i, string(p))
					e = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", time.Now())))
					if e != nil {
						log.Println(e)
						break loop
					}
				}
			}
		},
	)

	s.Handler = mux

	s.ListenAndServe()
}
