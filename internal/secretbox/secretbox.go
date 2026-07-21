// Package secretbox encrypts deployment-owned tenant secrets.
package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrInvalidKey = errors.New("invalid secret encryption key")

// ParseKey accepts the documented base64 key and a raw 32-byte value for
// local tests. Production configuration should use base64.
func ParseKey(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, ErrInvalidKey
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if len(value) == 32 {
		return []byte(value), nil
	}
	return nil, ErrInvalidKey
}

func NewKeyBase64() string {
	var key [32]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(key[:])
}

func Encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	if len(key) != 32 {
		return nil, nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("secret cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("secret gcm: %w", err)
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, nil), nonce, nil
}

func Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secret cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secret gcm: %w", err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, ErrInvalidKey
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func Last4(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 4 {
		return value
	}
	return value[len(value)-4:]
}

func Fingerprint(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
