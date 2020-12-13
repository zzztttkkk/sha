package sha

import (
	"testing"
)

func TestServer_ListenAndServeTLS(t *testing.T) {
	s := Default(nil)

	s.Handler = RequestHandlerFunc(
		func(ctx *RequestCtx) {
			_, _ = ctx.WriteString("Hello tls")
		},
	)

	s.ListenAndServeTLS("./tls/ztk.local+3.pem", "./tls/ztk.local+3-key.pem")
}
