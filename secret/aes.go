package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

var bufPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

func AesEncrypt(format string, args ...interface{}) (string, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	_, _ = fmt.Fprintf(buf, format, args...)

	plain, _ := ioutil.ReadAll(buf)
	for i := aes.BlockSize - len(plain)%aes.BlockSize; i > 0 && i != aes.BlockSize; i-- {
		plain = append(plain, ' ')
	}

	block, err := aes.NewCipher(gSecretKey)
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

var _AesDecryptError = errors.New("suna.secret: decrypt value error")

func AesDecrypt(f string) (string, error) {
	ciphertext, err := hex.DecodeString(f)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(gSecretKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return "", _AesDecryptError
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return "", _AesDecryptError
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return strings.TrimRight(string(ciphertext), " "), nil
}
