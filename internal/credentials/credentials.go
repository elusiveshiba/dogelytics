// Package credentials implements credential generation and one-way storage.
package credentials

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

const APIKeySHA256Prefix = "sha256:"

// GenerateID returns a URL-safe identifier containing n random bytes.
func GenerateID(n int) (string, error) {
	value := make([]byte, n)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

// HashPassword creates the password representation stored in PostgreSQL.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// HashAPIKeySecret creates the constant-time API-key representation stored in PostgreSQL.
func HashAPIKeySecret(secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return APIKeySHA256Prefix + base64.RawURLEncoding.EncodeToString(digest[:])
}
