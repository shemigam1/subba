package auth

import (
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correcthorsebattery")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("hash not argon2id-encoded: %q", hash)
	}
	if strings.Contains(hash, "correcthorsebattery") {
		t.Fatal("hash contains the plaintext password")
	}

	ok, err := VerifyPassword(hash, "correcthorsebattery")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("correct password did not verify")
	}

	ok, err = VerifyPassword(hash, "wrongpassword")
	if err != nil {
		t.Fatalf("VerifyPassword(wrong): %v", err)
	}
	if ok {
		t.Fatal("wrong password verified")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	a, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	b, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two hashes of the same password are identical — salt is not random")
	}
}

func TestVerifyPasswordMalformedHash(t *testing.T) {
	for _, encoded := range []string{
		"",
		"plaintext",
		"$bcrypt$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA", // wrong algorithm
		"$argon2id$v=19$m=65536,t=3,p=2$!!notb64!!$aGFzaA",
	} {
		if ok, err := VerifyPassword(encoded, "whatever"); err == nil || ok {
			t.Errorf("VerifyPassword(%q) = (%v, %v), want error and false", encoded, ok, err)
		}
	}
}

func TestRandomToken(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tok := RandomToken()
		// 32 bytes raw-URL-base64 encode to 43 chars.
		if len(tok) != 43 {
			t.Fatalf("token length = %d, want 43: %q", len(tok), tok)
		}
		if strings.ContainsAny(tok, "+/=") {
			t.Fatalf("token is not URL-safe: %q", tok)
		}
		if seen[tok] {
			t.Fatalf("duplicate token generated: %q", tok)
		}
		seen[tok] = true
	}
}

func TestHashToken(t *testing.T) {
	h1 := HashToken("some-token")
	h2 := HashToken("some-token")
	if h1 != h2 {
		t.Fatal("HashToken is not deterministic")
	}
	// hex SHA-256 is 64 chars.
	if len(h1) != 64 {
		t.Fatalf("hash length = %d, want 64", len(h1))
	}
	if h1 == HashToken("other-token") {
		t.Fatal("different tokens produced the same hash")
	}
	if strings.Contains(h1, "some-token") {
		t.Fatal("hash contains the plaintext token")
	}
}
