// Package crypto provides simple encrypt-at-rest for individual sensitive
// DB fields (currently just Child.EncryptedPesel) -- AES-256-GCM with a
// random nonce per encryption, base64-encoded for storage in a text
// column. This is the first encrypted-field feature in this codebase;
// nothing here is reused elsewhere yet.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/Johncode248/Rodzic-Chat/variables"
)

var ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")

func newGCM() (cipher.AEAD, error) {
	block, err := aes.NewCipher([]byte(variables.FieldEncryptionKey))
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptField returns a base64 string safe to store in a text column.
// Empty input returns empty output (no encryption of "no value").
func EncryptField(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptField reverses EncryptField. Empty input returns empty output.
func DecryptField(stored string) (string, error) {
	if stored == "" {
		return "", nil
	}

	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
