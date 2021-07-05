package session

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/sha/auth"
	"testing"
)

type IntSubject int64

func (sub IntSubject) GetID() string { return fmt.Sprintf("%d", sub) }

func (sub IntSubject) Info(ctx context.Context) interface{} { return nil }

type SessionReq struct {
	session   []byte
	sessionOk bool
}

func (sreq *SessionReq) UserAgent() string {
	return "go"
}

func (sreq *SessionReq) GetSessionID() *[]byte { return &sreq.session }

func (sreq *SessionReq) SetSessionID() { sreq.sessionOk = true }

func init() {
	opt := &Options{}
	opt.Redis.Addrs = []string{"127.0.0.1:16379"}
	Init(opt)
	auth.Init(auth.ManagerFunc(func(ctx context.Context) (auth.Subject, error) { return IntSubject(20), nil }))
}

func TestNew(t *testing.T) {
	req := &SessionReq{}
	s, e := New(context.Background(), req)
	fmt.Println(string(s), e, req.sessionOk)
	_ = s.Set(context.Background(), "qwer", 34)
	_, _ = s.Incr(context.Background(), "asdf", 1)
	fmt.Println(s.GetAll(context.Background()))
}
