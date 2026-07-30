package main

import (
	"github.com/zippyra/backend/shared/crypto"
)

func EncryptTOTPSecret(secret string) ([]byte, error) {
	return crypto.Encrypt([]byte(secret), "")
}

func DecryptTOTPSecret(ciphertext []byte) (string, error) {
	dec, err := crypto.Decrypt(ciphertext, "")
	if err != nil {
		return "", err
	}
	return string(dec), nil
}
