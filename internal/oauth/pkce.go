package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCE holds a code_verifier and S256 code_challenge pair.
type PKCE struct {
	Verifier  string
	Challenge string
	Method    string // always "S256"
}

// NewPKCE generates a high-entropy code_verifier and S256 challenge (RFC 7636).
func NewPKCE() (PKCE, error) {
	// 32 bytes -> 43 char base64url (within 43-128)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return PKCE{}, fmt.Errorf("pkce: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return PKCE{Verifier: verifier, Challenge: challenge, Method: "S256"}, nil
}

// RandomState returns an opaque CSRF state value.
func RandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
