package sha

import (
	"embed"
	"io"
	"time"

	"github.com/zzztttkkk/sha/utils"
)

func NewEmbedFSHandler(fs *embed.FS, modTime time.Time, pathRewrite func(ctx *RequestCtx) string) RequestHandler {
	var mt time.Time
	if modTime.IsZero() {
		mt = time.Now()
	}

	if pathRewrite == nil {
		pathRewrite = func(ctx *RequestCtx) string {
			path, _ := ctx.Request.URL.Params.Get("filepath")
			if len(path) < 1 {
				ctx.Response.statusCode = StatusNotFound
				return ""
			}
			return utils.S(path)
		}
	}

	return RequestHandlerFunc(func(ctx *RequestCtx) {
		w := &ctx.Response
		r := &ctx.Request

		path := pathRewrite(ctx)
		if len(path) < 1 {
			return
		}

		f, e := fs.Open(path)
		if e != nil {
			w.statusCode = _FSErrToHTTPError(e)
			return
		}
		stat, _ := f.Stat()

		setLastModified(w, mt)
		done, rangeReq := checkPreconditions(w, r, mt)
		if done {
			w.header.SetContentLength(0)
			return
		}

		setContentTypeForFile(w, path)

		if len(rangeReq) > 0 {
			w.statusCode = StatusBadRequest
			return
		}

		if r._method != _MHead {
			_, _ = io.CopyN(w, f, stat.Size())
		}
	})
}
