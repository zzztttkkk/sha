package sha

import (
	"crypto/tls"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"net"
)

type Session struct {
	conn  net.Conn
	isTLS bool
}

func NewSession(address string) (*Session, error) {
	conn, err := net.Dial("tcp4", address)
	if err != nil {
		return nil, err
	}
	return &Session{conn: conn}, nil
}

func NewSessionTLS(address string, conf *tls.Config) (*Session, error) {
	conn, err := tls.Dial("tcp4", address, conf)
	if err != nil {
		return nil, err
	}
	return &Session{conn: conn, isTLS: true}, nil
}

func (s *Session) Close() error { return s.conn.Close() }

var requestSendBufferPool = utils.NewBufferPoll(4096)
var responseReadBufferPool = utils.NewFixedSizeBufferPoll(4096, 8192)

func (s *Session) Send(ctx *RequestCtx) error {
	ctx.isTLS = s.isTLS

	sBuf := requestSendBufferPool.Get()
	defer requestSendBufferPool.Put(sBuf)

	req := &ctx.Request

	_, _ = sBuf.Write(req.Method)
	_ = sBuf.WriteByte(' ')
	_, _ = sBuf.Write(req.Path)
	sBuf.WriteString(" HTTP/1.1")
	sBuf.WriteString("\r\n")

	bodySize := len(ctx.buf)
	if bodySize > 0 {
		req.Header.SetContentLength(int64(bodySize))
	}

	req.Header.EachItem(func(item *utils.KvItem) bool {
		_, _ = sBuf.Write(item.Key)
		sBuf.WriteString(": ")
		_, _ = sBuf.Write(item.Val)
		sBuf.WriteString("\r\n")
		return true
	})
	sBuf.WriteString("\r\n")
	if bodySize > 0 {
		_, _ = sBuf.Write(ctx.buf)
	}

	_, err := s.conn.Write(sBuf.Data)
	if err != nil {
		return err
	}

	rBuf := responseReadBufferPool.Get()
	defer responseReadBufferPool.Put(rBuf)

	for {
		l, err := s.conn.Read(rBuf.Data)
		if err != nil {
			return err
		}

		if l < 1 {
			continue
		}

		// todo
		fmt.Println(string(rBuf.Data[:l]))
		return nil
	}
}
