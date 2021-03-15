package sha

//func TestFs(t *testing.T) {
//	server := Default(nil)
//	mux := NewMux(nil, nil)
//
//	mux.FilePath(
//		http.Dir("./"),
//		"get",
//		"/sha/filename:*",
//		true,
//		MiddlewareFunc(
//			func(ctx *RequestCtx, next func()) {
//				ctx.AutoCompress()
//				next()
//			},
//		),
//	)
//
//	server.Handler = mux
//	server.ListenAndServe()
//}
