package upload

// service_admin.go exposes admin-facing CRUD facade for Provider / Bucket / UploadKey / Rule
// plus a Provider connectivity probe used by the storage_admin handler.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/models"
)

// AdminAPI groups the management-plane operations the API handler needs.
type AdminAPI interface {
	// Providers
	ListProviders(ctx context.Context, tenantID string) ([]models.StorageProvider, error)
	GetProvider(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageProvider, error)
	CreateProvider(ctx context.Context, tenantID string, input ProviderSaveInput) (*models.StorageProvider, error)
	UpdateProvider(ctx context.Context, tenantID string, id uuid.UUID, input ProviderSaveInput) (*models.StorageProvider, error)
	DeleteProvider(ctx context.Context, tenantID string, id uuid.UUID) error
	TestProvider(ctx context.Context, tenantID string, id uuid.UUID) (*ProviderTestResult, error)

	// Buckets
	ListBuckets(ctx context.Context, tenantID string, providerID *uuid.UUID) ([]models.StorageBucket, error)
	GetBucket(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageBucket, error)
	CreateBucket(ctx context.Context, tenantID string, input BucketSaveInput) (*models.StorageBucket, error)
	UpdateBucket(ctx context.Context, tenantID string, id uuid.UUID, input BucketSaveInput) (*models.StorageBucket, error)
	DeleteBucket(ctx context.Context, tenantID string, id uuid.UUID) error

	// UploadKeys
	ListUploadKeys(ctx context.Context, tenantID string, bucketID *uuid.UUID) ([]models.UploadKey, error)
	GetUploadKey(ctx context.Context, tenantID string, id uuid.UUID) (*models.UploadKey, []models.UploadKeyRule, error)
	CreateUploadKey(ctx context.Context, tenantID string, input UploadKeySaveInput) (*models.UploadKey, error)
	UpdateUploadKey(ctx context.Context, tenantID string, id uuid.UUID, input UploadKeySaveInput) (*models.UploadKey, error)
	DeleteUploadKey(ctx context.Context, tenantID string, id uuid.UUID) error

	// Rules
	ListRulesByUploadKey(ctx context.Context, tenantID string, uploadKeyID uuid.UUID) ([]models.UploadKeyRule, error)
	CreateRule(ctx context.Context, tenantID string, uploadKeyID uuid.UUID, input UploadRuleSaveInput) (*models.UploadKeyRule, error)
	UpdateRule(ctx context.Context, tenantID string, id uuid.UUID, input UploadRuleSaveInput) (*models.UploadKeyRule, error)
	DeleteRule(ctx context.Context, tenantID string, id uuid.UUID) error

	// Mask helper exposed for tests / handlers.
	MaskSecret(value string) string
}

type ProviderSaveInput struct {
	ProviderKey string
	Name        string
	Driver      string
	Endpoint    string
	Region      string
	BaseURL     string
	AccessKey   string
	SecretKey   string
	IsDefault   bool
	Status      string
	Extra       models.MetaJSON
}

type ProviderTestResult struct {
	OK        bool
	Message   string
	LatencyMs int64
}

type BucketSaveInput struct {
	ProviderID    uuid.UUID
	BucketKey     string
	Name          string
	BucketName    string
	BasePath      string
	PublicBaseURL string
	IsPublic      bool
	Status        string
	Extra         models.MetaJSON
}

type UploadKeySaveInput struct {
	BucketID                 uuid.UUID
	Key                      string
	Name                     string
	PathTemplate             string
	DefaultRuleKey           string
	MaxSizeBytes             int64
	AllowedMimeTypes         models.StringList
	UploadMode               string
	IsFrontendVisible        bool
	PermissionKey            string
	FallbackKey              string
	ClientAccept             models.StringList
	DirectSizeThresholdBytes int64
	ExtraSchema              models.MetaJSON
	Visibility               string
	Status                   string
	Meta                     models.MetaJSON
}

type UploadRuleSaveInput struct {
	RuleKey            string
	Name               string
	SubPath            string
	FilenameStrategy   string
	MaxSizeBytes       int64
	AllowedMimeTypes   models.StringList
	ProcessPipeline    models.StringList
	ModeOverride       string
	VisibilityOverride string
	ClientAccept       models.StringList
	ExtraSchema        models.MetaJSON
	IsDefault          bool
	Status             string
	Meta               models.MetaJSON
}

// Admin returns the service cast to AdminAPI. Service implements AdminAPI directly.
func (s *service) Admin() AdminAPI { return s }

// MaskSecret exposes the repository-level secret mask helper.
func (s *service) MaskSecret(value string) string { return maskSecretValue(value) }

// ---------- providers ----------

func (s *service) ListProviders(ctx context.Context, tenantID string) ([]models.StorageProvider, error) {
	items, err := s.repo.ListProviders(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].AccessKeyEncrypted = maskSecretValue(items[i].AccessKeyEncrypted)
		items[i].SecretKeyEncrypted = maskSecretValue(items[i].SecretKeyEncrypted)
	}
	return items, nil
}

func (s *service) GetProvider(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageProvider, error) {
	item, err := s.repo.GetProviderByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	item.AccessKeyEncrypted = maskSecretValue(item.AccessKeyEncrypted)
	item.SecretKeyEncrypted = maskSecretValue(item.SecretKeyEncrypted)
	return item, nil
}

func (s *service) CreateProvider(ctx context.Context, tenantID string, input ProviderSaveInput) (*models.StorageProvider, error) {
	if strings.TrimSpace(input.ProviderKey) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Driver) == "" {
		return nil, errors.New("provider_key/name/driver are required")
	}
	provider := &models.StorageProvider{
		TenantID:           normalizeTenantID(tenantID),
		ProviderKey:        strings.TrimSpace(input.ProviderKey),
		Name:               strings.TrimSpace(input.Name),
		Driver:             strings.TrimSpace(input.Driver),
		Endpoint:           input.Endpoint,
		Region:             input.Region,
		BaseURL:            input.BaseURL,
		AccessKeyEncrypted: input.AccessKey,
		SecretKeyEncrypted: input.SecretKey,
		Extra:              input.Extra,
		IsDefault:          input.IsDefault,
		Status:             firstNonEmpty(input.Status, models.UploadProviderStatusReady),
	}
	if err := s.repo.EnsureProvider(ctx, provider); err != nil {
		return nil, err
	}
	return s.GetProvider(ctx, tenantID, provider.ID)
}

func (s *service) UpdateProvider(ctx context.Context, tenantID string, id uuid.UUID, input ProviderSaveInput) (*models.StorageProvider, error) {
	updated, err := s.repo.UpdateProviderByID(ctx, tenantID, id, ProviderUpdate{
		ProviderKey: input.ProviderKey,
		Name:        input.Name,
		Driver:      input.Driver,
		Endpoint:    input.Endpoint,
		Region:      input.Region,
		BaseURL:     input.BaseURL,
		AccessKey:   input.AccessKey,
		SecretKey:   input.SecretKey,
		IsDefault:   input.IsDefault,
		Status:      input.Status,
		Extra:       input.Extra,
	})
	if err != nil {
		return nil, err
	}
	updated.AccessKeyEncrypted = maskSecretValue(updated.AccessKeyEncrypted)
	updated.SecretKeyEncrypted = maskSecretValue(updated.SecretKeyEncrypted)
	return updated, nil
}

func (s *service) DeleteProvider(ctx context.Context, tenantID string, id uuid.UUID) error {
	buckets, err := s.repo.ListBuckets(ctx, tenantID, &id)
	if err != nil {
		return err
	}
	if len(buckets) > 0 {
		return fmt.Errorf("该 Provider 下仍有 %d 个 Bucket，请先删除或迁移", len(buckets))
	}
	return s.repo.DeleteProvider(ctx, tenantID, id)
}

func (s *service) TestProvider(ctx context.Context, tenantID string, id uuid.UUID) (*ProviderTestResult, error) {
	provider, err := s.repo.GetProviderDecryptedByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	bucket := models.StorageBucket{
		TenantID:   provider.TenantID,
		ProviderID: provider.ID,
	}
	// pick a representative bucket to make HealthCheck meaningful (e.g. resolve OSS endpoint)
	if buckets, listErr := s.repo.ListBuckets(ctx, tenantID, &provider.ID); listErr == nil && len(buckets) > 0 {
		bucket = buckets[0]
	}
	driver, err := s.registry.Open(DriverFactoryInput{
		Config:   s.cfg,
		Provider: *provider,
		Bucket:   bucket,
		Logger:   s.logger,
	})
	if err != nil {
		return &ProviderTestResult{OK: false, Message: err.Error()}, nil
	}
	start := time.Now()
	hcErr := driver.HealthCheck(ctx)
	latency := time.Since(start).Milliseconds()

	result := &ProviderTestResult{LatencyMs: latency}
	if hcErr != nil {
		result.OK = false
		result.Message = hcErr.Error()
	} else {
		result.OK = true
		result.Message = fmt.Sprintf("driver=%s ok", driver.Name())
	}

	outcome := audit.OutcomeSuccess
	if !result.OK {
		outcome = audit.OutcomeError
	}
	s.repo.recordAudit(ctx, audit.Event{
		Action:       "system.upload.provider.test",
		ResourceType: "upload_provider",
		ResourceID:   provider.ProviderKey,
		Outcome:      outcome,
		Metadata: map[string]any{
			"ok":         result.OK,
			"latency_ms": result.LatencyMs,
			"message":    result.Message,
		},
	})

	return result, nil
}

// ---------- buckets ----------

func (s *service) ListBuckets(ctx context.Context, tenantID string, providerID *uuid.UUID) ([]models.StorageBucket, error) {
	return s.repo.ListBuckets(ctx, tenantID, providerID)
}

func (s *service) GetBucket(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageBucket, error) {
	return s.repo.GetBucketByID(ctx, tenantID, id)
}

func (s *service) CreateBucket(ctx context.Context, tenantID string, input BucketSaveInput) (*models.StorageBucket, error) {
	if input.ProviderID == uuid.Nil || strings.TrimSpace(input.BucketKey) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.BucketName) == "" {
		return nil, errors.New("provider_id/bucket_key/name/bucket_name are required")
	}
	bucket := &models.StorageBucket{
		TenantID:      normalizeTenantID(tenantID),
		ProviderID:    input.ProviderID,
		BucketKey:     strings.TrimSpace(input.BucketKey),
		Name:          strings.TrimSpace(input.Name),
		BucketName:    strings.TrimSpace(input.BucketName),
		BasePath:      input.BasePath,
		PublicBaseURL: input.PublicBaseURL,
		IsPublic:      input.IsPublic,
		Status:        firstNonEmpty(input.Status, models.UploadProviderStatusReady),
		Extra:         input.Extra,
	}
	if err := s.repo.EnsureBucket(ctx, bucket); err != nil {
		return nil, err
	}
	return s.repo.GetBucketByID(ctx, tenantID, bucket.ID)
}

func (s *service) UpdateBucket(ctx context.Context, tenantID string, id uuid.UUID, input BucketSaveInput) (*models.StorageBucket, error) {
	return s.repo.UpdateBucketByID(ctx, tenantID, id, BucketUpdate{
		ProviderID:    input.ProviderID,
		BucketKey:     input.BucketKey,
		Name:          input.Name,
		BucketName:    input.BucketName,
		BasePath:      input.BasePath,
		PublicBaseURL: input.PublicBaseURL,
		IsPublic:      input.IsPublic,
		Status:        input.Status,
		Extra:         input.Extra,
	})
}

func (s *service) DeleteBucket(ctx context.Context, tenantID string, id uuid.UUID) error {
	keys, err := s.repo.ListUploadKeys(ctx, tenantID, &id)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return fmt.Errorf("该 Bucket 下仍有 %d 个 UploadKey，请先删除或迁移", len(keys))
	}
	return s.repo.DeleteBucket(ctx, tenantID, id)
}

// ---------- upload keys ----------

func (s *service) ListUploadKeys(ctx context.Context, tenantID string, bucketID *uuid.UUID) ([]models.UploadKey, error) {
	return s.repo.ListUploadKeys(ctx, tenantID, bucketID)
}

func (s *service) GetUploadKey(ctx context.Context, tenantID string, id uuid.UUID) (*models.UploadKey, []models.UploadKeyRule, error) {
	item, err := s.repo.GetUploadKeyByID(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	rules, err := s.repo.ListRulesByUploadKey(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	return item, rules, nil
}

func (s *service) CreateUploadKey(ctx context.Context, tenantID string, input UploadKeySaveInput) (*models.UploadKey, error) {
	if input.BucketID == uuid.Nil || strings.TrimSpace(input.Key) == "" || strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("bucket_id/key/name are required")
	}
	extraSchema, err := NormalizeUploadKeyExtraSchema(input.ExtraSchema)
	if err != nil {
		return nil, err
	}
	allowed := input.AllowedMimeTypes
	if allowed == nil {
		allowed = models.StringList{}
	}
	clientAccept := input.ClientAccept
	if clientAccept == nil {
		clientAccept = models.StringList{}
	}
	item := &models.UploadKey{
		TenantID:                 normalizeTenantID(tenantID),
		BucketID:                 input.BucketID,
		Key:                      strings.TrimSpace(input.Key),
		Name:                     strings.TrimSpace(input.Name),
		PathTemplate:             input.PathTemplate,
		DefaultRuleKey:           strings.TrimSpace(input.DefaultRuleKey),
		MaxSizeBytes:             input.MaxSizeBytes,
		AllowedMimeTypes:         allowed,
		UploadMode:               firstNonEmpty(input.UploadMode, models.UploadModeAuto),
		IsFrontendVisible:        input.IsFrontendVisible,
		PermissionKey:            strings.TrimSpace(input.PermissionKey),
		FallbackKey:              strings.TrimSpace(input.FallbackKey),
		ClientAccept:             clientAccept,
		DirectSizeThresholdBytes: input.DirectSizeThresholdBytes,
		ExtraSchema:              extraSchema,
		Visibility:               firstNonEmpty(input.Visibility, "public"),
		Status:                   firstNonEmpty(input.Status, models.UploadProviderStatusReady),
		Meta:                     input.Meta,
	}
	if err := s.repo.EnsureUploadKey(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetUploadKeyByID(ctx, tenantID, item.ID)
}

func (s *service) UpdateUploadKey(ctx context.Context, tenantID string, id uuid.UUID, input UploadKeySaveInput) (*models.UploadKey, error) {
	extraSchema, err := NormalizeUploadKeyExtraSchema(input.ExtraSchema)
	if err != nil {
		return nil, err
	}
	return s.repo.UpdateUploadKeyByID(ctx, tenantID, id, UploadKeyUpdate{
		BucketID:                 input.BucketID,
		Key:                      input.Key,
		Name:                     input.Name,
		PathTemplate:             input.PathTemplate,
		DefaultRuleKey:           input.DefaultRuleKey,
		MaxSizeBytes:             input.MaxSizeBytes,
		AllowedMimeTypes:         input.AllowedMimeTypes,
		UploadMode:               input.UploadMode,
		IsFrontendVisible:        input.IsFrontendVisible,
		PermissionKey:            input.PermissionKey,
		FallbackKey:              input.FallbackKey,
		ClientAccept:             input.ClientAccept,
		DirectSizeThresholdBytes: input.DirectSizeThresholdBytes,
		ExtraSchema:              extraSchema,
		Visibility:               input.Visibility,
		Status:                   input.Status,
		Meta:                     input.Meta,
	})
}

func (s *service) DeleteUploadKey(ctx context.Context, tenantID string, id uuid.UUID) error {
	rules, err := s.repo.ListRulesByUploadKey(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		return fmt.Errorf("该 UploadKey 下仍有 %d 条 Rule，请先删除", len(rules))
	}
	return s.repo.DeleteUploadKey(ctx, tenantID, id)
}

// ---------- rules ----------

func (s *service) ListRulesByUploadKey(ctx context.Context, tenantID string, uploadKeyID uuid.UUID) ([]models.UploadKeyRule, error) {
	return s.repo.ListRulesByUploadKey(ctx, tenantID, uploadKeyID)
}

func (s *service) CreateRule(ctx context.Context, tenantID string, uploadKeyID uuid.UUID, input UploadRuleSaveInput) (*models.UploadKeyRule, error) {
	if uploadKeyID == uuid.Nil || strings.TrimSpace(input.RuleKey) == "" || strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("upload_key_id/rule_key/name are required")
	}
	extraSchema, err := NormalizeUploadRuleExtraSchema(input.ExtraSchema)
	if err != nil {
		return nil, err
	}
	allowed := input.AllowedMimeTypes
	if allowed == nil {
		allowed = models.StringList{}
	}
	pipeline := input.ProcessPipeline
	if pipeline == nil {
		pipeline = models.StringList{}
	}
	clientAccept := input.ClientAccept
	if clientAccept == nil {
		clientAccept = models.StringList{}
	}
	item := &models.UploadKeyRule{
		TenantID:           normalizeTenantID(tenantID),
		UploadKeyID:        uploadKeyID,
		RuleKey:            strings.TrimSpace(input.RuleKey),
		Name:               strings.TrimSpace(input.Name),
		SubPath:            input.SubPath,
		FilenameStrategy:   firstNonEmpty(input.FilenameStrategy, "uuid"),
		MaxSizeBytes:       input.MaxSizeBytes,
		AllowedMimeTypes:   allowed,
		ProcessPipeline:    pipeline,
		ModeOverride:       firstNonEmpty(input.ModeOverride, models.UploadModeInherit),
		VisibilityOverride: firstNonEmpty(input.VisibilityOverride, models.VisibilityOverrideInherit),
		ClientAccept:       clientAccept,
		ExtraSchema:        extraSchema,
		IsDefault:          input.IsDefault,
		Status:             firstNonEmpty(input.Status, models.UploadProviderStatusReady),
		Meta:               input.Meta,
	}
	if err := s.repo.EnsureUploadRule(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetUploadRuleByID(ctx, tenantID, item.ID)
}

func (s *service) UpdateRule(ctx context.Context, tenantID string, id uuid.UUID, input UploadRuleSaveInput) (*models.UploadKeyRule, error) {
	extraSchema, err := NormalizeUploadRuleExtraSchema(input.ExtraSchema)
	if err != nil {
		return nil, err
	}
	return s.repo.UpdateUploadRuleByID(ctx, tenantID, id, UploadRuleUpdate{
		RuleKey:            input.RuleKey,
		Name:               input.Name,
		SubPath:            input.SubPath,
		FilenameStrategy:   input.FilenameStrategy,
		MaxSizeBytes:       input.MaxSizeBytes,
		AllowedMimeTypes:   input.AllowedMimeTypes,
		ProcessPipeline:    input.ProcessPipeline,
		ModeOverride:       input.ModeOverride,
		VisibilityOverride: input.VisibilityOverride,
		ClientAccept:       input.ClientAccept,
		ExtraSchema:        extraSchema,
		IsDefault:          input.IsDefault,
		Status:             input.Status,
		Meta:               input.Meta,
	})
}

func (s *service) DeleteRule(ctx context.Context, tenantID string, id uuid.UUID) error {
	return s.repo.DeleteUploadRule(ctx, tenantID, id)
}
