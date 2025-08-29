package utils

import (
	"os"
	"testing"

	dbModels "github.com/ZiplEix/scrabble/api/models/database"
)

func TestGenerateToken_FailsWithoutSecret(t *testing.T) {
	// Ensure JWT_SECRET unset
	old := os.Getenv("JWT_SECRET")
	_ = os.Unsetenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", old)

	// getJWTSecret panics when not set → GenerateToken should panic through SignedString? No: getJWTSecret called during signing.
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when JWT_SECRET is missing")
		}
	}()
	_, _ = GenerateToken(dbModels.User{ID: 1, Username: "bob"})
}

func TestGenerateToken_Success(t *testing.T) {
	old := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Setenv("JWT_SECRET", old)

	tok, err := GenerateToken(dbModels.User{ID: 7, Username: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok == "" {
		t.Fatalf("expected non-empty token")
	}
}
