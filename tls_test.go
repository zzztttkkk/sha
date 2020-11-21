package suna

import (
	"context"
	"testing"
)

func TestServer_ListenAndServeTLS(t *testing.T) {
	s := Server{
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("Hello tls")
		},
	)

	s.ListenAndServeTLS("./tls/ztk.local+4.pem", "./tls/ztk.local+4-key.pem")
}
