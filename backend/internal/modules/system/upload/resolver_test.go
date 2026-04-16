package upload

import (
	"context"
	"testing"

	"github.com/maben/backend/internal/modules/system/models"
)

func TestResolveEffectiveConfigParsesRuleSuffixAndMergesRuleFields(t *testing.T) {
	repo, _ := newUploadTestRepository(t)
	_, _, uploadKey, _ := seedUploadHierarchy(t, repo, "tenant-a")

	avatarRule := models.UploadKeyRule{
		TenantID:         "tenant-a",
		UploadKeyID:      uploadKey.ID,
		RuleKey:          "avatar",
		Name:             "Avatar Rule",
		SubPath:          "avatars",
		FilenameStrategy: "original",
		MaxSizeBytes:     512 * 1024,
		AllowedMimeTypes: models.StringList{"image/png"},
		ProcessPipeline:  models.StringList{"watermark"},
		Status:           models.UploadProviderStatusReady,
	}
	if err := repo.EnsureUploadRule(context.Background(), &avatarRule); err != nil {
		t.Fatalf("EnsureUploadRule(avatar) error = %v", err)
	}

	effective, err := repo.ResolveEffectiveConfig(context.Background(), "tenant-a", ResolveConfigInput{
		Key:      "media.default.avatar",
		Fallback: true,
	})
	if err != nil {
		t.Fatalf("ResolveEffectiveConfig() error = %v", err)
	}
	if effective.ResolvedKey != "media.default" {
		t.Fatalf("ResolvedKey = %q", effective.ResolvedKey)
	}
	if effective.ResolvedRule != "avatar" {
		t.Fatalf("ResolvedRule = %q", effective.ResolvedRule)
	}
	if effective.MaxSizeBytes != 512*1024 {
		t.Fatalf("MaxSizeBytes = %d", effective.MaxSizeBytes)
	}
	if effective.SubPath != "avatars" {
		t.Fatalf("SubPath = %q", effective.SubPath)
	}
	if effective.FilenameStrategy != "original" {
		t.Fatalf("FilenameStrategy = %q", effective.FilenameStrategy)
	}
	if len(effective.AllowedMimeTypes) != 1 || effective.AllowedMimeTypes[0] != "image/png" {
		t.Fatalf("AllowedMimeTypes = %#v", effective.AllowedMimeTypes)
	}
	if len(effective.ProcessPipeline) != 1 || effective.ProcessPipeline[0] != "watermark" {
		t.Fatalf("ProcessPipeline = %#v", effective.ProcessPipeline)
	}
}

func TestResolveEffectiveConfigFallbackToDefault(t *testing.T) {
	repo, _ := newUploadTestRepository(t)
	_, _, uploadKey, rule := seedUploadHierarchy(t, repo, "tenant-a")

	effective, err := repo.ResolveEffectiveConfig(context.Background(), "tenant-a", ResolveConfigInput{
		Key:      "missing.key",
		Fallback: true,
	})
	if err != nil {
		t.Fatalf("ResolveEffectiveConfig() fallback error = %v", err)
	}
	if !effective.FallbackUsed {
		t.Fatalf("FallbackUsed = false, want true")
	}
	if effective.ResolvedKey != uploadKey.Key {
		t.Fatalf("ResolvedKey = %q, want %q", effective.ResolvedKey, uploadKey.Key)
	}
	if effective.ResolvedRule != rule.RuleKey {
		t.Fatalf("ResolvedRule = %q, want %q", effective.ResolvedRule, rule.RuleKey)
	}
}

func TestResolveEffectiveConfigWithoutFallbackReturnsError(t *testing.T) {
	repo, _ := newUploadTestRepository(t)
	seedUploadHierarchy(t, repo, "tenant-a")

	_, err := repo.ResolveEffectiveConfig(context.Background(), "tenant-a", ResolveConfigInput{
		Key:      "missing.key",
		Fallback: false,
	})
	if err == nil {
		t.Fatalf("ResolveEffectiveConfig() should fail when fallback disabled")
	}
	if err != ErrUploadKeyNotFound {
		t.Fatalf("ResolveEffectiveConfig() error = %v, want %v", err, ErrUploadKeyNotFound)
	}
}

func TestParseUploadKey(t *testing.T) {
	parsed := ParseUploadKey(" media.default.avatar ")
	if parsed.Raw != "media.default.avatar" {
		t.Fatalf("Raw = %q", parsed.Raw)
	}
	if parsed.Key != "media.default.avatar" {
		t.Fatalf("Key = %q", parsed.Key)
	}
	if parsed.Rule != "" {
		t.Fatalf("Rule = %q, want empty", parsed.Rule)
	}
}

