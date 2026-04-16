package upload

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maben/backend/internal/config"
)

func TestSecretCipherEncryptDecrypt(t *testing.T) {
	cipher, err := NewSecretCipher(config.UploadConfig{
		SecretMasterKeys:   []string{"local:test-master-key"},
		SecretCurrentKeyID: "local",
		SecretCacheTTL:     60,
	})
	if err != nil {
		t.Fatalf("NewSecretCipher() error = %v", err)
	}

	encrypted, err := cipher.Encrypt(context.Background(), "super-secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted == "super-secret" {
		t.Fatalf("Encrypt() returned plaintext")
	}

	plaintext, err := cipher.Decrypt(context.Background(), encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "super-secret" {
		t.Fatalf("Decrypt() = %q, want %q", plaintext, "super-secret")
	}
}

func TestSecretCipherRejectsMalformedPayload(t *testing.T) {
	cipher, err := NewSecretCipher(config.UploadConfig{
		SecretMasterKeys:   []string{"local:test-master-key"},
		SecretCurrentKeyID: "local",
	})
	if err != nil {
		t.Fatalf("NewSecretCipher() error = %v", err)
	}

	_, err = cipher.Decrypt(context.Background(), "bad-payload")
	if !errors.Is(err, ErrSecretCipherMalformed) {
		t.Fatalf("Decrypt() error = %v, want %v", err, ErrSecretCipherMalformed)
	}
}

func TestNewSecretCipherRequiresKey(t *testing.T) {
	_, err := NewSecretCipher(config.UploadConfig{})
	if !errors.Is(err, ErrSecretCipherUnavailable) {
		t.Fatalf("NewSecretCipher() error = %v, want %v", err, ErrSecretCipherUnavailable)
	}
}

func TestSecretCipherCurrentKeyAndCacheHelpers(t *testing.T) {
	cipher, err := NewSecretCipher(config.UploadConfig{
		SecretMasterKeys:   []string{"local:test-master-key"},
		SecretCurrentKeyID: "local",
		SecretCacheTTL:     60,
	})
	if err != nil {
		t.Fatalf("NewSecretCipher() error = %v", err)
	}
	if cipher.CurrentKeyID() != "local" {
		t.Fatalf("CurrentKeyID() = %q, want %q", cipher.CurrentKeyID(), "local")
	}

	cache := newSecretCache(int(time.Minute.Seconds()), 16)
	cache.Set("payload", "plaintext")
	if got, ok := cache.Get("payload"); !ok || got != "plaintext" {
		t.Fatalf("secret cache Get() = (%q,%v), want (%q,true)", got, ok, "plaintext")
	}
}

func TestSecretCacheSetUpdatesAndEvicts(t *testing.T) {
	cache := newSecretCache(60, 1)
	cache.Set("first", "1")
	cache.Set("first", "2")
	if got, ok := cache.Get("first"); !ok || got != "2" {
		t.Fatalf("secret cache update = (%q,%v), want (%q,true)", got, ok, "2")
	}
	cache.Set("second", "3")
	if _, ok := cache.Get("first"); ok {
		t.Fatalf("secret cache should evict oldest entry when over capacity")
	}
}

