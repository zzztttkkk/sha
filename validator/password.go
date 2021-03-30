package validator

import (
	"github.com/zzztttkkk/sha/utils"
	"golang.org/x/crypto/bcrypt"
	"unicode"
)

type BcryptPassword []byte

var BcryptPasswordValidator func(data []byte) bool
var BcryptPasswordCost = bcrypt.DefaultCost

func defaultPwdValidator(data []byte) bool {
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

func (p *BcryptPassword) FormValue(data []byte) bool {
	var ok bool
	if BcryptPasswordValidator != nil {
		ok = BcryptPasswordValidator(data)
	} else {
		ok = defaultPwdValidator(data)
	}
	if !ok {
		return false
	}
	h, e := bcrypt.GenerateFromPassword(data, BcryptPasswordCost)
	if e != nil {
		return false
	}
	*p = append(*p, h...)
	return true
}
