package validator

import (
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

func (p *Password) FromBytes(data []byte) bool {
	if !PasswordValidator(data) {
		return false
	}
	*p = append(*p, data...)
	return true
}

func (p *Password) BcryptHash(cost int) ([]byte, error) { return bcrypt.GenerateFromPassword(*p, cost) }

func (p *Password) MatchTo(hash []byte) bool { return bcrypt.CompareHashAndPassword(hash, *p) == nil }
