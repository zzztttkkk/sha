package secret

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

func JwtEncode(data jwt.Claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	ts, err := token.SignedString(secretKey)
	if err != nil {
		panic(err)
	}
	return ts
}

var signMethodError = fmt.Errorf("snow.secret.jwt: unexpected signing method")
var JwtInvalidError = fmt.Errorf("snow.secret.jwt: invalid data")

func JwtDecode(ts string) (jwt.Claims, error) {
	token, err := jwt.Parse(ts, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, signMethodError
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return token.Claims, JwtInvalidError
	}
	return token.Claims, nil
}
