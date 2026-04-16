package upload

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

var ErrUploadKeyNotFound = errors.New("upload key not found")

type ResolveConfigInput struct {
	Key      string
	Rule     string
	Fallback bool
}

type ParsedUploadKey struct {
	Raw  string
	Key  string
	Rule string
}

type EffectiveConfig struct {
	Provider                 models.StorageProvider
	Bucket                   models.StorageBucket
	UploadKey                models.UploadKey
	Rule                     models.UploadKeyRule
	ResolvedKey              string
	ResolvedRule             string
	FallbackUsed             bool
	MaxSizeBytes             int64
	AllowedMimeTypes         models.StringList
	PathTemplate             string
	SubPath                  string
	FilenameStrategy         string
	ProcessPipeline          models.StringList
	UploadMode               string
	IsFrontendVisible        bool
	PermissionKey            string
	FallbackKey              string
	ClientAccept             models.StringList
	DirectSizeThresholdBytes int64
	Visibility               string
}

func ParseUploadKey(value string) ParsedUploadKey {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ParsedUploadKey{}
	}
	return ParsedUploadKey{
		Raw: trimmed,
		Key: trimmed,
	}
}

func (r *Repository) ResolveEffectiveConfig(ctx context.Context, tenantID string, input ResolveConfigInput) (*EffectiveConfig, error) {
	tenantID = normalizeTenantID(tenantID)
	key := strings.TrimSpace(input.Key)
	rule := strings.TrimSpace(input.Rule)
	fallbackAllowed := input.Fallback
	if key == "" {
		fallbackAllowed = true
	}

	parsed := ParseUploadKey(key)
	cfg, selector, err := r.resolveConfig(ctx, tenantID, parsed)
	if err != nil && fallbackAllowed && errors.Is(err, ErrUploadKeyNotFound) {
		cfg, err = r.GetDefaultConfig(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		return buildEffectiveConfig(cfg, "", "", true), nil
	}
	if err != nil {
		return nil, err
	}

	resolvedRule := rule
	if resolvedRule == "" {
		resolvedRule = selector.Rule
	}
	if resolvedRule != "" {
		exactRule, exactErr := r.GetUploadRuleByKey(ctx, tenantID, cfg.UploadKey.ID, resolvedRule)
		if exactErr == nil {
			cfg.Rule = *exactRule
		} else if errors.Is(exactErr, gorm.ErrRecordNotFound) {
			resolvedRule = ""
		} else if !errors.Is(exactErr, gorm.ErrRecordNotFound) {
			return nil, exactErr
		}
	}
	return buildEffectiveConfig(cfg, selector.Key, resolvedRule, false), nil
}

func (r *Repository) resolveConfig(ctx context.Context, tenantID string, parsed ParsedUploadKey) (*ResolvedConfig, ParsedUploadKey, error) {
	if parsed.Key == "" {
		return nil, ParsedUploadKey{}, ErrUploadKeyNotFound
	}
	cfg, err := r.GetResolvedConfigByKey(ctx, tenantID, parsed.Key)
	if err == nil {
		return cfg, parsed, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ParsedUploadKey{}, err
	}

	lastDot := strings.LastIndex(parsed.Key, ".")
	if lastDot <= 0 || lastDot >= len(parsed.Key)-1 {
		return nil, ParsedUploadKey{}, ErrUploadKeyNotFound
	}
	candidate := ParsedUploadKey{
		Raw:  parsed.Raw,
		Key:  parsed.Key[:lastDot],
		Rule: parsed.Key[lastDot+1:],
	}
	cfg, err = r.GetResolvedConfigByKey(ctx, tenantID, candidate.Key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ParsedUploadKey{}, ErrUploadKeyNotFound
		}
		return nil, ParsedUploadKey{}, err
	}
	return cfg, candidate, nil
}

func buildEffectiveConfig(cfg *ResolvedConfig, resolvedKey, resolvedRule string, fallbackUsed bool) *EffectiveConfig {
	if cfg == nil {
		return nil
	}
	rule := cfg.Rule
	maxSize := rule.MaxSizeBytes
	if maxSize <= 0 || cfg.UploadKey.MaxSizeBytes > 0 && cfg.UploadKey.MaxSizeBytes < maxSize {
		if cfg.UploadKey.MaxSizeBytes > 0 {
			maxSize = cfg.UploadKey.MaxSizeBytes
		}
	}
	if maxSize <= 0 {
		maxSize = rule.MaxSizeBytes
	}
	allowedMimeTypes := rule.AllowedMimeTypes
	if len(allowedMimeTypes) == 0 {
		allowedMimeTypes = cfg.UploadKey.AllowedMimeTypes
	}
	processPipeline := cfg.Rule.ProcessPipeline
	if len(processPipeline) == 0 {
		processPipeline = models.StringList{}
	}
	clientAccept := cfg.Rule.ClientAccept
	if len(clientAccept) == 0 {
		clientAccept = cfg.UploadKey.ClientAccept
	}
	if len(clientAccept) == 0 {
		clientAccept = models.StringList{}
	}
	ruleKey := strings.TrimSpace(resolvedRule)
	if ruleKey == "" {
		ruleKey = rule.RuleKey
	}
	key := strings.TrimSpace(resolvedKey)
	if key == "" {
		key = cfg.UploadKey.Key
	}
	uploadMode := strings.TrimSpace(cfg.UploadKey.UploadMode)
	if uploadMode == "" {
		uploadMode = models.UploadModeAuto
	}
	if override := strings.TrimSpace(cfg.Rule.ModeOverride); override != "" && override != models.UploadModeInherit {
		uploadMode = override
	}
	visibility := strings.TrimSpace(cfg.UploadKey.Visibility)
	if visibility == "" {
		visibility = "public"
	}
	if override := strings.TrimSpace(cfg.Rule.VisibilityOverride); override != "" && override != models.VisibilityOverrideInherit {
		visibility = override
	}

	return &EffectiveConfig{
		Provider:                 cfg.Provider,
		Bucket:                   cfg.Bucket,
		UploadKey:                cfg.UploadKey,
		Rule:                     rule,
		ResolvedKey:              key,
		ResolvedRule:             ruleKey,
		FallbackUsed:             fallbackUsed,
		MaxSizeBytes:             maxSize,
		AllowedMimeTypes:         allowedMimeTypes,
		PathTemplate:             cfg.UploadKey.PathTemplate,
		SubPath:                  rule.SubPath,
		FilenameStrategy:         rule.FilenameStrategy,
		ProcessPipeline:          processPipeline,
		UploadMode:               uploadMode,
		IsFrontendVisible:        cfg.UploadKey.IsFrontendVisible,
		PermissionKey:            strings.TrimSpace(cfg.UploadKey.PermissionKey),
		FallbackKey:              strings.TrimSpace(cfg.UploadKey.FallbackKey),
		ClientAccept:             clientAccept,
		DirectSizeThresholdBytes: cfg.UploadKey.DirectSizeThresholdBytes,
		Visibility:               visibility,
	}
}

func (r *Repository) GetUploadRuleByKey(ctx context.Context, tenantID string, uploadKeyID uuid.UUID, ruleKey string) (*models.UploadKeyRule, error) {
	tenantID = normalizeTenantID(tenantID)
	ruleKey = strings.TrimSpace(ruleKey)
	if ruleKey == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var rule models.UploadKeyRule
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND upload_key_id = ? AND rule_key = ? AND deleted_at IS NULL", tenantID, uploadKeyID, ruleKey).
		First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}
