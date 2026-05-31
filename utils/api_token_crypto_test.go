package utils

import (
	"encoding/base64"
	"testing"
)

const testEncryptionKey = "12345678901234567890123456789012"
const testHashSecret = "12345678901234567890123456789012"

func TestHashAPITokenRequiresSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "")

	_, err := HashAPIToken("token")
	if err == nil {
		t.Fatal("expected missing secret error")
	}
}

func TestHashAPITokenIsStable(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", testHashSecret)

	first, err := HashAPIToken("token")
	if err != nil {
		t.Fatalf("hash api token: %v", err)
	}
	second, err := HashAPIToken("token")
	if err != nil {
		t.Fatalf("hash api token again: %v", err)
	}

	if first == "" {
		t.Fatal("expected hash")
	}
	if first != second {
		t.Fatalf("hash mismatch: %q != %q", first, second)
	}
}

func TestEncryptDecryptAPITokenRoundTrip(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", testEncryptionKey)

	encrypted, err := EncryptAPIToken("api-token")
	if err != nil {
		t.Fatalf("encrypt api token: %v", err)
	}
	if encrypted == "api-token" {
		t.Fatal("expected encrypted token to differ from plaintext")
	}

	decrypted, err := DecryptAPIToken(encrypted)
	if err != nil {
		t.Fatalf("decrypt api token: %v", err)
	}
	if decrypted != "api-token" {
		t.Fatalf("decrypted token = %q, want %q", decrypted, "api-token")
	}
}

func TestEncryptAPITokenRejectsMissingKey(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "")

	_, err := EncryptAPIToken("api-token")
	if err == nil {
		t.Fatal("expected missing encryption key error")
	}
}

func TestEncryptAPITokenRejectsInvalidKey(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "short-key")

	_, err := EncryptAPIToken("api-token")
	if err == nil {
		t.Fatal("expected invalid encryption key error")
	}
}

func TestEncryptionKeyAcceptsBase64Encodings(t *testing.T) {
	rawKey := []byte(testEncryptionKey)

	tests := []struct {
		name string
		key  string
	}{
		{
			name: "base64 raw URL encoding",
			key:  base64.RawURLEncoding.EncodeToString(rawKey),
		},
		{
			name: "standard base64 encoding",
			key:  base64.StdEncoding.EncodeToString(rawKey),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("API_TOKEN_ENCRYPTION_KEY", tt.key)

			got, err := apiTokenEncryptionKey()
			if err != nil {
				t.Fatalf("api token encryption key: %v", err)
			}
			if string(got) != testEncryptionKey {
				t.Fatalf("decoded key = %q, want %q", string(got), testEncryptionKey)
			}
		})
	}
}

func TestDecryptAPITokenRejectsInvalidInput(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", testEncryptionKey)

	tests := []struct {
		name      string
		encrypted string
	}{
		{
			name:      "invalid base64",
			encrypted: "%",
		},
		{
			name:      "too short",
			encrypted: base64.RawURLEncoding.EncodeToString([]byte("short")),
		},
		{
			name:      "bad ciphertext",
			encrypted: base64.RawURLEncoding.EncodeToString(append(make([]byte, apiTokenNonceSize), []byte("bad-ciphertext")...)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptAPIToken(tt.encrypted)
			if err == nil {
				t.Fatal("expected decrypt error")
			}
		})
	}
}
