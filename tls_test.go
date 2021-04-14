package sha

import (
	"testing"
)

func TestServer_ListenAndServeTLS(t *testing.T) {
	conf := ServerOptions{}
	conf.TLS.Key = "./tls/sha.local-key.pem"
	conf.TLS.Cert = "./tls/sha.local.pem"

	s := New(nil, nil, &conf)

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("Hello tls")
		},
	)

	s.ListenAndServe()
}
