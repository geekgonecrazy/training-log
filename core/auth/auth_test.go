package auth

import (
	"errors"
	"testing"
	"time"
)

func TestPasswordRoundtrip(t *testing.T) {
	hash, err := HashPassword("hunter2-correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyPassword(hash, "hunter2-correct-horse"); err != nil {
		t.Fatalf("expected match, got %v", err)
	}
	if err := VerifyPassword(hash, "wrong"); !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("expected mismatch, got %v", err)
	}
}

func TestJWTRoundtrip(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()
	tok, err := IssueAccessToken(secret, 42, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseAccessToken(secret, tok, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != 42 {
		t.Fatalf("subject = %d, want 42", claims.Subject)
	}
}

func TestJWTExpiry(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()
	tok, err := IssueAccessToken(secret, 1, time.Second, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseAccessToken(secret, tok, now.Add(2*time.Second)); !errors.Is(err, ErrJWTExpired) {
		t.Fatalf("expected expired, got %v", err)
	}
}

func TestJWTBadSig(t *testing.T) {
	secret := []byte("test-secret")
	tok, _ := IssueAccessToken(secret, 1, time.Minute, time.Now())
	if _, err := ParseAccessToken([]byte("other"), tok, time.Now()); !errors.Is(err, ErrJWTBadSig) {
		t.Fatalf("expected bad sig, got %v", err)
	}
}

func TestRefreshTokenHash(t *testing.T) {
	tok, hash, err := GenerateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if HashRefreshToken(tok) != hash {
		t.Fatal("hash mismatch")
	}
}
