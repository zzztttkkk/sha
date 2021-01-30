package sha

import (
	"testing"
)

func TestServer_ListenAndServeTLS(t *testing.T) {
	conf := ServerConf{}
	conf.Tls.Key = "./tls/sha.local-key.pem"
	conf.Tls.Cert = "./tls/sha.local.pem"

	s := New(nil, &conf, nil, nil)

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("Hello tls")
		},
	)

	s.ListenAndServe()
}
