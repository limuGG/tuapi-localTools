package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type DecryptRequest struct {
	Secret     string `json:"secret"`
	Ciphertext string `json:"ciphertext"`
}

type EncryptRequest struct {
	Secret string          `json:"secret"`
	Plain  json.RawMessage `json:"plain"`
}

// AESCBCDecrypt hex decodes a piece of data and then decrypts it using CBC mode.
func AESCBCDecrypt(cipherKey, ciphertext []byte) ([]byte, error) {
	dst := make([]byte, hex.DecodedLen(len(ciphertext)))
	if _, err := hex.Decode(dst, bytes.ToLower(ciphertext)); err != nil {
		return nil, err
	}
	ciphertext = dst
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := ciphertext
	mode.CryptBlocks(plaintext, ciphertext)
	return pkcs5UnPadding(plaintext)
}

// AESCBCEncrypt uses CBC mode to encrypt a piece of data and then encodes it in hex.
func AESCBCEncrypt(cipherKey, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	plaintext = pkcs5Padding(plaintext, blockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	dst := make([]byte, hex.EncodedLen(len(ciphertext)))
	hex.Encode(dst, ciphertext)
	return bytes.ToUpper(dst), nil
}

func DecryptHandler() http.HandlerFunc { return JSONHandlerWrap(decryptFunc) }

func EncryptHandler() http.HandlerFunc { return JSONHandlerWrap(encryptFunc) }

func decryptFunc(req *DecryptRequest) (json.RawMessage, error) {
	if len(req.Secret) < 32 {
		return nil, ErrorParams("密钥长度不足32位")
	}
	if len(req.Ciphertext) == 0 {
		return nil, ErrorParams("密文不能为空")
	}
	result, err := AESCBCDecrypt([]byte(req.Secret)[:32], []byte(req.Ciphertext))
	if err != nil {
		return nil, ErrorParams("解密失败,请检查参数")
	}
	return result, nil
}

func encryptFunc(req *EncryptRequest) (string, error) {
	if len(req.Secret) < 32 {
		return "", ErrorParams("密钥长度不足32位")
	}
	if len(req.Plain) == 0 {
		return "", ErrorParams("明文不能为空")
	}
	result, err := AESCBCEncrypt([]byte(req.Secret)[:32], req.Plain)
	if err != nil {
		return "", ErrorParams("加密失败,请检查参数")
	}
	return string(result), nil
}

func pkcs5Padding(plaintext []byte, blockSize int) []byte {
	n := byte(blockSize - len(plaintext)%blockSize)
	for i := byte(0); i < n; i++ {
		plaintext = append(plaintext, n)
	}
	return plaintext
}

func pkcs5UnPadding(r []byte) ([]byte, error) {
	l := len(r)
	if l == 0 {
		return nil, errors.New("input padded bytes is empty")
	}
	last := int(r[l-1])
	if l-last < 0 {
		return nil, errors.New("input padded bytes is invalid")
	}
	n := byte(last)
	pad := r[l-last : l]
	isPad := true
	for _, v := range pad {
		if v != n {
			isPad = false
			break
		}
	}
	if !isPad {
		return nil, errors.New("remove pad error")
	}
	return r[:l-last], nil
}
