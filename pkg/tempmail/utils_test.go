package tempmail

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	charset := "abc123"
	length := 10

	result := GenerateRandomString(length, charset)

	if len(result) != length {
		t.Errorf("expected length %d, got %d", length, len(result))
	}

	for _, c := range result {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("character %c not in charset", c)
		}
	}
}

func TestGenerateUsername(t *testing.T) {
	length := 10
	result := GenerateUsername(length)

	if len(result) != length {
		t.Errorf("expected length %d, got %d", length, len(result))
	}

	// Should only contain lowercase and digits
	for _, c := range result {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			t.Errorf("invalid character %c in username", c)
		}
	}
}

func TestGeneratePassword(t *testing.T) {
	length := 16
	result := GeneratePassword(length)

	if len(result) != length {
		t.Errorf("expected length %d, got %d", length, len(result))
	}

	// Should contain alphanumeric characters
	for _, c := range result {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Errorf("invalid character %c in password", c)
		}
	}
}

func TestGenerateRandomStringUniqueness(t *testing.T) {
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	results := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result := GenerateRandomString(20, charset)
		if results[result] {
			t.Error("generated duplicate string")
		}
		results[result] = true
	}
}
