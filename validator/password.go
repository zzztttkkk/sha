package validator

import (
	"database/sql/driver"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"golang.org/x/crypto/bcrypt"
	"unicode"
)

type Password []byte

var PasswordValidator func(data []byte) bool = func(data []byte) bool {
	s := utils.S(data)
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 6 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		default:
			return false
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

var ErrBadPassword = fmt.Errorf("validator: bad passowrd format")

func (p *Password) FromBytes(data []byte) error {
	if !PasswordValidator(data) {
		return ErrBadPassword
	}
	*p = append(*p, data...)
	return nil
}

func (p *Password) Validate() error {
	if !PasswordValidator(*p) {
		return ErrBadPassword
	}
	return nil
}

func (p *Password) BcryptHash(cost int) ([]byte, error) { return bcrypt.GenerateFromPassword(*p, cost) }

func (p *Password) MatchTo(hash []byte) bool { return bcrypt.CompareHashAndPassword(hash, *p) == nil }

func (p *Password) Scan(v interface{}) error {
	switch rv := v.(type) {
	case []byte:
		*p = rv
		return nil
	case string:
		*p = utils.B(rv)
		return nil
	default:
		return ErrBadPassword
	}
}

func (p *Password) Value() (driver.Value, error) {
	return []byte(*p), nil
}

var _ Field = (*Password)(nil)
