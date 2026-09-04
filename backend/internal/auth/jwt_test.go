package auth

import (
	"testing"
	"time"
)

func TestIssueAndParseToken(t *testing.T) {
	tm := NewTokenManager("test-secret", time.Hour)

	token, exp, err := tm.Issue("user-123", "alice")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if exp.IsZero() {
		t.Fatal("zero expiry")
	}

	claims, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("expected user-123, got %s", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Fatalf("expected alice, got %s", claims.Username)
	}
}

func TestParseExpiredToken(t *testing.T) {
	tm := NewTokenManager("test-secret", -time.Hour)
	token, _, err := tm.Issue("user-123", "alice")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if _, err := tm.Parse(token); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	tm1 := NewTokenManager("secret-one", time.Hour)
	token, _, err := tm1.Issue("user-123", "alice")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	tm2 := NewTokenManager("secret-two", time.Hour)
	if _, err := tm2.Parse(token); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
