package internal

const (
	RCtxKeySession  = ".suna.s"
	RCtxKeyRemoteIp = ".suna.r"
	RCtxKeyUser     = ".suna.u"
)

type StdCtxKey int

var (
	RCtxKeyStdCtx = StdCtxKey(0)
)
