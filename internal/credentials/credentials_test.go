package credentials

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestCredentialRepresentations(t *testing.T) {
	id, err := GenerateID(16)
	if err != nil || len(id) < 20 {
		t.Fatalf("generate id: value=%q err=%v", id, err)
	}
	passwordHash, err := HashPassword("a-long-test-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("a-long-test-password")); err != nil {
		t.Fatalf("verify password hash: %v", err)
	}
	keyHash := HashAPIKeySecret("secret")
	if !strings.HasPrefix(keyHash, APIKeySHA256Prefix) || strings.Contains(keyHash, "secret") {
		t.Fatalf("unexpected API-key hash: %q", keyHash)
	}
}
