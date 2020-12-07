// copied from `go.stdlib: http/fs.go`

package suna

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP file system request Handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

//goland:noinspection GoImportUsedAsName
func dirList(ctx *RequestCtx, f http.File) {
	res := &ctx.Response

	dirs, err := f.Readdir(-1)
	if err != nil {
		res.statusCode = StatusInternalServerError
		return
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })

	res.Header.SetContentType(MIMEHtml)
	_, _ = fmt.Fprintf(res, "<pre>\n")
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: name}
		fmt.Fprintf(res, "<a href=\"%s\">%s</a>\n", url.String(), htmlReplacer.Replace(name))
	}
	fmt.Fprintf(res, "</pre>\n")
}

// errNoOverlap is returned by serveContent's parseRange if first-byte-pos of
// all of the byte-range-spec values is greater than the content size.
var errNoOverlap = errors.New("invalid range: failed to overlap")

// if name is empty, filename is unknown. (used for mime type, before sniffing)
// if modtime.IsZero(), modtime is unknown.
// content must be seeked to the beginning of the file.
// The sizeFunc is called at most once. Its error, if any, is sent in the HTTP response.
func serveContent(ctx *RequestCtx, name string, modtime time.Time, sizeFunc func() (int64, error), content io.ReadSeeker) {
	w := &ctx.Response
	r := &ctx.Request

	setLastModified(w, modtime)
	done, rangeReq := checkPreconditions(w, r, modtime)
	if done {
		w.Header.SetContentLength(0)
		return
	}

	// If Content-Type isn't set, use the file's extension to find it, but
	// if the Content-Type is unset explicitly, do not sniff the type.
	_, haveType := w.Header.Get(internal.B(HeaderContentType))
	var ctype string
	if !haveType {
		ctype = mime.TypeByExtension(filepath.Ext(name))
		if ctype != "" {
			w.Header.SetContentType(internal.B(ctype))
		} else {
			//todo guess
		}
	}

	size, err := sizeFunc()
	if err != nil {
		w.statusCode = StatusInternalServerError
		return
	}

	// handle Content-Range header.
	sendSize := size
	var sendContent io.Reader = content
	if size >= 0 {
		ranges, err := parseRange(rangeReq, size)
		if err != nil {
			if err == errNoOverlap {
				w.Header.Set(internal.B(HeaderContentRange), internal.B(fmt.Sprintf("bytes */%d", size)))
			}
			w.statusCode = StatusRequestedRangeNotSatisfiable
			return
		}
		if sumRangesSize(ranges) > size {
			// The total number of bytes in all the ranges
			// is larger than the size of the file by
			// itself, so this is probably an attack, or a
			// dumb client. Ignore the range request.
			ranges = nil
		}
		switch {
		case len(ranges) == 1:
			// RFC 7233, Section 4.1:
			// "If a single part is being transferred, the server
			// generating the 206 response MUST generate a
			// Content-Range header field, describing what range
			// of the selected representation is enclosed, and a
			// payload consisting of the range.
			// ...
			// A server MUST NOT generate a multipart response to
			// a request for a single range, since a client that
			// does not request multiple parts might not support
			// multipart responses."
			ra := ranges[0]
			if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
				w.statusCode = StatusRequestedRangeNotSatisfiable
				return
			}
			sendSize = ra.length
			w.statusCode = StatusPartialContent
			w.Header.Set(internal.B(HeaderContentRange), ra.contentRange(size))
		case len(ranges) > 1:
			sendSize = rangesMIMESize(ranges, ctype, size)
			w.statusCode = StatusPartialContent

			pr, pw := io.Pipe()
			mw := multipart.NewWriter(pw)
			w.Header.Set(internal.B(HeaderContentType), internal.B("multipart/byteranges; boundary="+mw.Boundary()))
			sendContent = pr
			defer pr.Close() // cause writing goroutine to fail and exit if CopyN doesn't finish.
			go func() {
				for _, ra := range ranges {
					part, err := mw.CreatePart(ra.mimeHeader(ctype, size))
					if err != nil {
						_ = pw.CloseWithError(err)
						return
					}
					if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
						_ = pw.CloseWithError(err)
						return
					}
					if _, err := io.CopyN(part, content, ra.length); err != nil {
						_ = pw.CloseWithError(err)
						return
					}
				}
				_ = mw.Close()
				_ = pw.Close()
			}()
		}
		w.Header.Set(internal.B(HeaderAcceptRanges), []byte("bytes"))
	}

	if string(r.Method) != "HEAD" {
		w.Header.SetContentLength(sendSize)
		_, _ = io.CopyN(ctx, sendContent, sendSize)
	}
}

// scanETag determines if a syntactically valid ETag is present at s. If so,
// the ETag and remaining text after consuming ETag is returned. Otherwise,
// it returns "", "".
func scanETag(s string) (etag string, remain string) {
	s = textproto.TrimString(s)
	start := 0
	if strings.HasPrefix(s, "W/") {
		start = 2
	}
	if len(s[start:]) < 2 || s[start] != '"' {
		return "", ""
	}
	// ETag is either W/"text" or "text".
	// See RFC 7232 2.3.
	for i := start + 1; i < len(s); i++ {
		c := s[i]
		switch {
		// Character values allowed in ETags.
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
		case c == '"':
			return s[:i+1], s[i+1:]
		default:
			return "", ""
		}
	}
	return "", ""
}

// etagStrongMatch reports whether a and b match using strong ETag comparison.
// Assumes a and b are valid ETags.
func etagStrongMatch(a, b string) bool {
	return a == b && a != "" && a[0] == '"'
}

// etagWeakMatch reports whether a and b match using weak ETag comparison.
// Assumes a and b are valid ETags.
func etagWeakMatch(a, b string) bool {
	return strings.TrimPrefix(a, "W/") == strings.TrimPrefix(b, "W/")
}

// condResult is the result of an HTTP request precondition check.
// See https://tools.ietf.org/html/rfc7232 section 3.
type condResult int

const (
	condNone condResult = iota
	condTrue
	condFalse
)

var headerIfMatch = []byte("If-Match")
var headerEtag = []byte("Etag")

func checkIfMatch(w *Response, r *Request) condResult {
	imb, _ := r.Header.Get(headerIfMatch)
	if len(imb) < 1 {
		return condNone
	}
	im := string(imb)
	for {
		im = textproto.TrimString(im)
		if len(im) == 0 {
			break
		}
		if im[0] == ',' {
			im = im[1:]
			continue
		}
		if im[0] == '*' {
			return condTrue
		}
		etag, remain := scanETag(im)
		if etag == "" {
			break
		}

		etagV, _ := w.Header.Get(headerEtag)
		if etagStrongMatch(etag, internal.S(etagV)) {
			return condTrue
		}
		im = remain
	}

	return condFalse
}

func checkIfUnmodifiedSince(r *Request, modtime time.Time) condResult {
	ius, _ := r.Header.Get(internal.B(HeaderIfUnmodifiedSince))
	if len(ius) < 1 || isZeroTime(modtime) {
		return condNone
	}

	t, err := http.ParseTime(internal.S(ius))
	if err != nil {
		return condNone
	}

	// The Last-Modified header truncates sub-second precision so
	// the modtime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if modtime.Before(t) || modtime.Equal(t) {
		return condTrue
	}
	return condFalse
}

func checkIfNoneMatch(w *Response, r *Request) condResult {
	inmb, _ := r.Header.Get(internal.B(HeaderIfNoneMatch))
	if len(inmb) < 1 {
		return condNone
	}
	buf := string(inmb)
	for {
		buf = textproto.TrimString(buf)
		if len(buf) == 0 {
			break
		}
		if buf[0] == ',' {
			buf = buf[1:]
			continue
		}
		if buf[0] == '*' {
			return condFalse
		}
		etag, remain := scanETag(buf)
		if etag == "" {
			break
		}
		etagV, _ := w.Header.Get(headerEtag)
		if etagWeakMatch(etag, string(etagV)) {
			return condFalse
		}
		buf = remain
	}
	return condTrue
}

func checkIfModifiedSince(r *Request, modtime time.Time) condResult {
	m := internal.S(r.Method)
	if m != "GET" && m != "HEAD" {
		return condNone
	}

	ims, _ := r.Header.Get(internal.B(HeaderIfModifiedSince))
	if len(ims) < 1 || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(internal.S(ims))
	if err != nil {
		return condNone
	}
	// The Last-Modified header truncates sub-second precision so
	// the modtime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if modtime.Before(t) || modtime.Equal(t) {
		return condFalse
	}
	return condTrue
}

func checkIfRange(w *Response, r *Request, modtime time.Time) condResult {
	m := internal.S(r.Method)
	if m != "GET" && m != "HEAD" {
		return condNone
	}
	irb, _ := r.Header.Get(internal.B(HeaderIfRange))
	if len(irb) < 1 {
		return condNone
	}
	ir := string(irb)
	etag, _ := scanETag(ir)
	if etag != "" {
		etagV, _ := w.Header.Get(headerEtag)
		if etagStrongMatch(etag, string(etagV)) {
			return condTrue
		} else {
			return condFalse
		}
	}
	// The If-Range value is typically the ETag value, but it may also be
	// the modtime date. See golang.org/issue/8367.
	if modtime.IsZero() {
		return condFalse
	}
	t, err := http.ParseTime(ir)
	if err != nil {
		return condFalse
	}
	if t.Unix() == modtime.Unix() {
		return condTrue
	}
	return condFalse
}

var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

const fsTimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func setLastModified(w *Response, modtime time.Time) {
	if !isZeroTime(modtime) {
		w.Header.Set(internal.B(HeaderLastModified), internal.B(modtime.UTC().Format(fsTimeFormat)))
	}
}

func writeNotModified(w *Response) {
	// RFC 7232 section 4.1:
	// a sender SHOULD NOT generate representation metadata other than the
	// above listed fields unless said metadata exists for the purpose of
	// guiding cache updates (e.g., Last-Modified might be useful if the
	// response does not have an ETag field).
	w.statusCode = StatusNotModified
	w.Header.Del(internal.B(HeaderContentType))
	w.Header.Del(internal.B(HeaderContentLength))
	etagV, _ := w.Header.Get(headerEtag)
	if len(etagV) > 0 {
		w.Header.Del(internal.B(HeaderLastModified))
	}
}

// checkPreconditions evaluates request preconditions and reports whether a precondition
// resulted in sending StatusNotModified or StatusPreconditionFailed.
func checkPreconditions(w *Response, r *Request, modtime time.Time) (done bool, rangeHeader string) {
	// This function carefully follows RFC 7232 section 6.
	ch := checkIfMatch(w, r)
	if ch == condNone {
		ch = checkIfUnmodifiedSince(r, modtime)
	}
	if ch == condFalse {
		w.statusCode = StatusPreconditionFailed
		return true, ""
	}

	method := internal.S(r.Method)

	switch checkIfNoneMatch(w, r) {
	case condFalse:
		if method == "GET" || method == "HEAD" {
			writeNotModified(w)
			return true, ""
		} else {
			w.statusCode = StatusPreconditionFailed
			return true, ""
		}
	case condNone:
		if checkIfModifiedSince(r, modtime) == condFalse {
			writeNotModified(w)
			return true, ""
		}
	}

	rangeHeaderB, _ := r.Header.Get(internal.B(HeaderRange))
	if len(rangeHeaderB) > 0 && checkIfRange(w, r, modtime) == condFalse {
		rangeHeader = ""
	} else {
		rangeHeader = internal.S(rangeHeaderB)
	}
	return false, rangeHeader
}

var indexPage = []byte("/index.html")

// name is '/'-separated, not filepath.Separator.
func serveFile(ctx *RequestCtx, fs http.FileSystem, name string, index bool) {
	w := &ctx.Response
	r := &ctx.Request

	// redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	if bytes.HasSuffix(r.Path, indexPage) {
		localRedirect(w, r, "./")
		return
	}

	f, err := fs.Open(name)
	if err != nil {
		w.statusCode = toHTTPError(err)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		w.statusCode = toHTTPError(err)
		return
	}

	if d.IsDir() {
		urlV := internal.S(r.Path)
		// redirect if the directory name doesn't end in a slash
		if urlV == "" || urlV[len(urlV)-1] != '/' {
			localRedirect(w, r, path.Base(urlV)+"/")
			return
		}

		// use contents of index.html for directory, if present
		index := strings.TrimSuffix(name, "/") + internal.S(indexPage)
		ff, err := fs.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	// Still a directory? (we didn't find an index.html file)
	if d.IsDir() {
		if index {
			if checkIfModifiedSince(r, d.ModTime()) == condFalse {
				writeNotModified(w)
				return
			}
			setLastModified(w, d.ModTime())
			dirList(ctx, f)
			return
		}
		ctx.SetStatus(StatusNotFound)
		return
	}

	// serveContent will check modification time
	sizeFunc := func() (int64, error) { return d.Size(), nil }
	serveContent(ctx, d.Name(), d.ModTime(), sizeFunc, f)
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (httpStatus int) {
	if os.IsNotExist(err) {
		return StatusNotFound
	}
	if os.IsPermission(err) {
		return StatusForbidden
	}
	// Default:
	return StatusInternalServerError
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w *Response, r *Request, newPath string) {
	if q := internal.S(r.rawPath); len(q) > 0 {
		newPath += "?" + q
	}
	w.Header.Set(internal.B(HeaderLocation), internal.B(newPath))
	w.statusCode = StatusMovedPermanently
}

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) []byte {
	return internal.B(fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size))
}

func (r httpRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Range": {internal.S(r.contentRange(size))},
		"Content-Type":  {contentType},
	}
}

// parseRange parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}
		start, end := textproto.TrimString(ra[:i]), textproto.TrimString(ra[i+1:])
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, errNoOverlap
	}
	return ranges, nil
}

// countingWriter counts how many bytes have been written to it.
type countingWriter int64

func (w *countingWriter) Write(p []byte) (n int, err error) {
	*w += countingWriter(len(p))
	return len(p), nil
}

// rangesMIMESize returns the number of bytes it takes to encode the
// provided ranges as a multipart response.
func rangesMIMESize(ranges []httpRange, contentType string, contentSize int64) (encSize int64) {
	var w countingWriter
	mw := multipart.NewWriter(&w)
	for _, ra := range ranges {
		_, _ = mw.CreatePart(ra.mimeHeader(contentType, contentSize))
		encSize += ra.length
	}
	_ = mw.Close()
	encSize += int64(w)
	return
}

func sumRangesSize(ranges []httpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}
