package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	masterKey := "12345678901234567890123456789012"
	plaintext := []byte("secret_webhook_key_or_sap_credentials")

	encrypted, err := Encrypt(plaintext, masterKey)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(encrypted, plaintext) {
		t.Fatalf("Encrypted ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(encrypted, masterKey)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Expected decrypted %s, got %s", plaintext, decrypted)
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	masterKey := "12345678901234567890123456789012"
	plaintext := []byte("hello_world")

	encrypted, _ := Encrypt(plaintext, masterKey)
	encrypted[len(encrypted)-1] ^= 0xff // tamper last byte

	_, err := Decrypt(encrypted, masterKey)
	if err == nil {
		t.Fatalf("Expected error when decrypting tampered ciphertext, got nil")
	}
}
