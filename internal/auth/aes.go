package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func GetAESDecrypted(encrypted string) ([]byte, error) {
	key := GetRandomInfo() + printThis() + delMethod()
	iv := showMe() + parseInfo()

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher([]byte(key))

	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("block size cant be zero")
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, ciphertext)
	ciphertext = PKCS5UnPadding(ciphertext)

	return ciphertext, nil
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])

	return src[:(length - unpadding)]
}

func GetRandomInfo() string {
	return "WlAKYMpyhVCmR"
}

func printThis() string {
	return "VbjNVQnSqODbG"
}

func delMethod() string {
	return "dBJqRI"
}

func showMe() string {
	return "WviOsKg"
}

func parseInfo() string {
	return "UqGQsfzwJ"
}
