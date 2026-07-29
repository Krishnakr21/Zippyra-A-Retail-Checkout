package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var defaultDevKeyBase64 = "emlwcHlyYS1kZXYtdG90cC1lbmNyeXB0aW9uLWtleTMy" // base64 of "zippyra-dev-totp-encryption-key32"

func GetEncryptionKey(customKey string) ([]byte, error) {
	keyStr := customKey
	if keyStr == "" {
		keyStr = os.Getenv("ENCRYPTION_MASTER_KEY")
	}
	if keyStr == "" {
		keyStr = os.Getenv("ADMIN_TOTP_ENCRYPTION_KEY")
	}
	if keyStr == "" {
		keyStr = defaultDevKeyBase64
	}

	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil || len(key) != 32 {
		if len(keyStr) == 32 {
			return []byte(keyStr), nil
		}
		devBytes := []byte("zippyra-dev-totp-encryption-key32")
		return devBytes[:32], nil
	}
	return key, nil
}

func Encrypt(data []byte, customKey string) ([]byte, error) {
	key, err := GetEncryptionKey(customKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func Decrypt(ciphertext []byte, customKey string) ([]byte, error) {
	key, err := GetEncryptionKey(customKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
