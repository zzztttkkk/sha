package sha

import (
	"github.com/zzztttkkk/sha/utils"
	"mime"
	"path/filepath"
	"strings"
)

const (
	MIMEJson      = "application/json"
	MIMEForm      = "application/x-www-form-urlencoded"
	MIMEMultiPart = "multipart/form-data"
	MIMEText      = "text/plain"
	MIMEMarkdown  = "text/markdown"
	MIMEHtml      = "text/html"
	MIMEPng       = "image/png"
	MIMEJpeg      = "image/jpeg"
	MIMEUnknown   = "application/octet-stream"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
var defaultMIMEMap = map[string]string{
	".js":   "application/javascript",
	".css":  "text/css",
	".html": "text/html",

	".apng": "image/apng",
	".git":  "image/gif",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".ico":  "image/x-icon",
}

func setContentTypeForFile(res *Response, path string) string {
	bv, haveType := res.Header().Get(HeaderContentType)
	var ctype string
	if !haveType {
		ext := strings.ToLower(filepath.Ext(path))
		var ok bool
		ctype, ok = defaultMIMEMap[ext]
		if !ok {
			ctype = mime.TypeByExtension(ext)
		}
		if ctype == "" {
			ctype = MIMEUnknown
		}
		res.Header().SetContentType(ctype)
		return ctype
	}
	return utils.S(bv)
}
