package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	password := "super-secret-123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == password {
		t.Fatal("hash empty or equal to password")
	}

	ok, err := VerifyPassword(password, hash)
	if err != nil || !ok {
		t.Fatalf("expected verify to succeed, got ok=%v err=%v", ok, err)
	}

	ok, err = VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if ok {
		t.Fatal("expected verify to fail for wrong password")
	}
}

func TestVerifyPasswordMalformedHash(t *testing.T) {
	ok, err := VerifyPassword("pw", "not-a-valid-hash")
	if err == nil {
		t.Fatal("expected error for malformed hash")
	}
	if ok {
		t.Fatal("expected ok=false for malformed hash")
	}
}
