package sha

import (
	"embed"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

func NewEmbedFSHandler(fs *embed.FS, modTime time.Time, pathRewrite func(ctx *RequestCtx) string) RequestHandler {
	var mt time.Time
	if modTime.IsZero() {
		mt = time.Now()
	}

	if pathRewrite == nil {
		pathRewrite = func(ctx *RequestCtx) string {
			path, _ := ctx.Request.URL.Params.Get("filepath")
			if len(path) < 0 {
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

		_, haveType := w.Header().Get(HeaderContentType)
		var ctype string
		if !haveType {
			ext := strings.ToLower(filepath.Ext(path))
			var ok bool
			ctype, ok = defaultMIMEMap[ext]
			if !ok {
				ctype = mime.TypeByExtension(ext)
			}
			w.Header().SetContentType(ctype)
		}

		if len(rangeReq) > 0 {
			w.statusCode = StatusBadRequest
			return
		}

		if r._method != _MHead {
			_, _ = io.CopyN(w, f, stat.Size())
		}
	})
}
