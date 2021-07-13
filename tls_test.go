package sha

import (
	"context"
	"testing"
)

func TestServer_ListenAndServeTLS(t *testing.T) {
	conf := ServerOptions{}
	conf.TLS.Key = "./tls/sha.local-key.pem"
	conf.TLS.Cert = "./tls/sha.local.pem"

	s := New(context.Background(), nil, &conf)

	s.Handler = RequestCtxHandlerFunc(
		func(ctx *RequestCtx) {
			_ = ctx.WriteString("Hello tls")
		},
	)

	s.ListenAndServe()
}
