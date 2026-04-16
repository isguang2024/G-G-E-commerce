package upload

import (
	"context"
	"testing"
	"time"

	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/google/uuid"
)

func TestRepositoryProviderSecretHooksEncryptAndDecrypt(t *testing.T) {
	cipher, err := NewSecretCipher(config.UploadConfig{
		SecretMasterKeys:   []string{"local:test-master-key"},
		SecretCurrentKeyID: "local",
	})
	if err != nil {
		t.Fatalf("NewSecretCipher() error = %v", err)
	}

	repo := NewRepository(nil).WithSecretCipher(cipher)
	provider := &models.StorageProvider{
		AccessKeyEncrypted: "access-key",
		SecretKeyEncrypted: "secret-key",
	}

	if err := repo.encryptProviderSecrets(context.Background(), provider); err != nil {
		t.Fatalf("encryptProviderSecrets() error = %v", err)
	}
	if provider.AccessKeyEncrypted == "access-key" || provider.SecretKeyEncrypted == "secret-key" {
		t.Fatalf("encryptProviderSecrets() did not encrypt secrets")
	}
	if !isEncryptedSecret(provider.AccessKeyEncrypted) || !isEncryptedSecret(provider.SecretKeyEncrypted) {
		t.Fatalf("encryptProviderSecrets() produced unexpected payload format")
	}

	if err := repo.decryptProviderSecrets(context.Background(), provider); err != nil {
		t.Fatalf("decryptProviderSecrets() error = %v", err)
	}
	if provider.AccessKeyEncrypted != "access-key" {
		t.Fatalf("AccessKeyEncrypted = %q, want %q", provider.AccessKeyEncrypted, "access-key")
	}
	if provider.SecretKeyEncrypted != "secret-key" {
		t.Fatalf("SecretKeyEncrypted = %q, want %q", provider.SecretKeyEncrypted, "secret-key")
	}
}

func TestRepositoryProviderSecretHooksSkipPlaintextWhenCipherUnavailable(t *testing.T) {
	repo := NewRepository(nil)
	provider := &models.StorageProvider{
		AccessKeyEncrypted: "plain-access",
		SecretKeyEncrypted: "plain-secret",
	}

	if err := repo.encryptProviderSecrets(context.Background(), provider); err != nil {
		t.Fatalf("encryptProviderSecrets() error = %v", err)
	}
	if provider.AccessKeyEncrypted != "plain-access" || provider.SecretKeyEncrypted != "plain-secret" {
		t.Fatalf("encryptProviderSecrets() should keep plaintext when cipher is nil")
	}
}

func TestIsEncryptedSecret(t *testing.T) {
	if isEncryptedSecret("plain") {
		t.Fatalf("isEncryptedSecret() = true for plaintext")
	}
	if !isEncryptedSecret("gge:v1:local:payload") {
		t.Fatalf("isEncryptedSecret() = false for encrypted payload")
	}
}

func TestMaskSecretValue(t *testing.T) {
	if got := maskSecretValue(""); got != "" {
		t.Fatalf("maskSecretValue(\"\") = %q, want empty string", got)
	}
	if got := maskSecretValue("abcd"); got != "****" {
		t.Fatalf("maskSecretValue(\"abcd\") = %q, want %q", got, "****")
	}
	if got := maskSecretValue("abcdefghi"); got != "ab*****hi" {
		t.Fatalf("maskSecretValue(\"abcdefghi\") = %q, want %q", got, "ab*****hi")
	}
}

func TestRepositoryGetCachedUploadConfigUsesLocalCache(t *testing.T) {
	repo := NewRepository(nil).WithResolvedConfigCache(newLocalResolvedConfigCache(time.Minute))
	callCount := 0
	loader := func(context.Context) (*ResolvedConfig, error) {
		callCount++
		return &ResolvedConfig{
			UploadKey: models.UploadKey{Key: "media.default"},
		}, nil
	}

	first, err := repo.getCachedUploadConfig(context.Background(), "tenant-a", "media.default", loader)
	if err != nil {
		t.Fatalf("getCachedUploadConfig() first error = %v", err)
	}
	second, err := repo.getCachedUploadConfig(context.Background(), "tenant-a", "media.default", loader)
	if err != nil {
		t.Fatalf("getCachedUploadConfig() second error = %v", err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, want 1", callCount)
	}
	if first.UploadKey.Key != second.UploadKey.Key {
		t.Fatalf("cached upload key = %q, want %q", second.UploadKey.Key, first.UploadKey.Key)
	}
}

func TestRepositoryInvalidateUploadConfigCacheClearsCachedEntry(t *testing.T) {
	repo := NewRepository(nil).WithResolvedConfigCache(newLocalResolvedConfigCache(time.Minute))
	callCount := 0
	loader := func(context.Context) (*ResolvedConfig, error) {
		callCount++
		return &ResolvedConfig{
			UploadKey: models.UploadKey{ID: uuid.New(), Key: "media.default"},
		}, nil
	}

	if _, err := repo.getCachedUploadConfig(context.Background(), "tenant-a", "media.default", loader); err != nil {
		t.Fatalf("prime cache error = %v", err)
	}
	repo.invalidateUploadConfigCache(context.Background(), "tenant-a", "media.default", uuid.Nil)
	if _, err := repo.getCachedUploadConfig(context.Background(), "tenant-a", "media.default", loader); err != nil {
		t.Fatalf("reload cache error = %v", err)
	}
	if callCount != 2 {
		t.Fatalf("loader called %d times after invalidation, want 2", callCount)
	}
}

func TestRepositoryGetCachedUploadRuleUsesLocalCache(t *testing.T) {
	repo := NewRepository(nil).WithResolvedConfigCache(newLocalResolvedConfigCache(time.Minute))
	uploadKeyID := uuid.New()
	callCount := 0
	loader := func(context.Context) (*models.UploadKeyRule, error) {
		callCount++
		return &models.UploadKeyRule{
			UploadKeyID: uploadKeyID,
			RuleKey:     "image",
		}, nil
	}

	if _, err := repo.getCachedUploadRule(context.Background(), uploadKeyID, loader); err != nil {
		t.Fatalf("getCachedUploadRule() first error = %v", err)
	}
	rule, err := repo.getCachedUploadRule(context.Background(), uploadKeyID, loader)
	if err != nil {
		t.Fatalf("getCachedUploadRule() second error = %v", err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, want 1", callCount)
	}
	if rule.RuleKey != "image" {
		t.Fatalf("cached rule key = %q, want %q", rule.RuleKey, "image")
	}
}

func TestResolvedConfigCacheHandleInvalidationPayload(t *testing.T) {
	cache := newLocalResolvedConfigCache(time.Minute)
	cache.Set(context.Background(), uploadKeyCacheKey("tenant-a", "media.default"), &ResolvedConfig{
		UploadKey: models.UploadKey{Key: "media.default"},
	})
	cache.Set(context.Background(), uploadRuleCacheKey(uuid.Nil), &models.UploadKeyRule{RuleKey: "image"})

	cache.handleInvalidationPayload(`{"keys":["upload_rules:00000000-0000-0000-0000-000000000000"],"prefixes":["upload_key:tenant-a:"]}`)

	var cfg ResolvedConfig
	if cache.Get(context.Background(), uploadKeyCacheKey("tenant-a", "media.default"), &cfg) {
		t.Fatalf("upload key cache should be invalidated")
	}
	var rule models.UploadKeyRule
	if cache.Get(context.Background(), uploadRuleCacheKey(uuid.Nil), &rule) {
		t.Fatalf("upload rule cache should be invalidated")
	}
}

func TestNewLocalResolvedConfigCacheUsesDefaultTTL(t *testing.T) {
	cache := newLocalResolvedConfigCache(0)
	if cache.ttl != uploadConfigCacheTTLDefault {
		t.Fatalf("cache.ttl = %v, want %v", cache.ttl, uploadConfigCacheTTLDefault)
	}
}

