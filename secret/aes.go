package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

func AesEncrypt(data string) (string, error) {
	plain := []byte(data)
	for i := aes.BlockSize - len(plain)%aes.BlockSize; i > 0 && i != aes.BlockSize; i-- {
		plain = append(plain, ' ')
	}

	block, err := aes.NewCipher(_SecretKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plain)
	return hex.EncodeToString(ciphertext), nil
}

var ErrAesDecrypt = errors.New("suna.secret: decrypt value error")

func AesDecrypt(f string) (string, error) {
	ciphertext, err := hex.DecodeString(f)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(_SecretKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return "", ErrAesDecrypt
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return "", ErrAesDecrypt
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return strings.TrimRight(string(ciphertext), " "), nil
}
