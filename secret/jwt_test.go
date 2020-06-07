package secret

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

func TestJwt(t *testing.T) {
	m := jwt.MapClaims{
		"exp": time.Now().Unix(),
		"a":   "i can eat glass",
	}
	token := JwtEncode(m)
	time.Sleep(time.Second)
	v, e := JwtDecode(token)
	fmt.Println(v, e, token)
}
