package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// RandomToken returns a 32-byte cryptographically random, URL-safe token. Used for
// session ids, magic-link tokens, and API-key secrets. Panics if the system CSPRNG
// fails, since a silently zeroed buffer would make every token predictable.
func RandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("auth: crypto/rand unavailable: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashToken returns the hex SHA-256 of a token. We store only the hash for magic
// links and API keys, so a database leak never exposes a usable secret.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
