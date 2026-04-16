package main

import (
	"context"
	"testing"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/upload"
)

func TestUploadSecretKeyID(t *testing.T) {
	if got := uploadSecretKeyID("gge:v1:next:payload"); got != "next" {
		t.Fatalf("uploadSecretKeyID() = %q, want %q", got, "next")
	}
	if got := uploadSecretKeyID("plain"); got != "" {
		t.Fatalf("uploadSecretKeyID() = %q, want empty", got)
	}
}

func TestShouldRotateSecret(t *testing.T) {
	cipher, err := upload.NewSecretCipher(config.UploadConfig{
		SecretMasterKeys:   []string{"old:test-old-key", "next:test-next-key"},
		SecretCurrentKeyID: "next",
	})
	if err != nil {
		t.Fatalf("NewSecretCipher() error = %v", err)
	}

	currentEncrypted, err := cipher.Encrypt(context.Background(), "secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if shouldRotateSecret(cipher, currentEncrypted) {
		t.Fatalf("shouldRotateSecret() = true for current key")
	}
	if !shouldRotateSecret(cipher, "plain-secret") {
		t.Fatalf("shouldRotateSecret() = false for plaintext")
	}
	if !shouldRotateSecret(cipher, "gge:v1:old:payload") {
		t.Fatalf("shouldRotateSecret() = false for old key id")
	}
}
