package upload

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/models"
)

var defaultTenantID = "default"

type Repository struct {
	db            *gorm.DB
	cipher        SecretCipher
	auditRecorder audit.Recorder
	configCache   *resolvedConfigCache
	defaultKey    string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db:            db,
		auditRecorder: audit.Noop{},
	}
}

func (r *Repository) WithSecretCipher(cipher SecretCipher) *Repository {
	if r == nil {
		return nil
	}
	r.cipher = cipher
	return r
}

func (r *Repository) WithAuditRecorder(recorder audit.Recorder) *Repository {
	if r == nil {
		return nil
	}
	if recorder == nil {
		r.auditRecorder = audit.Noop{}
		return r
	}
	r.auditRecorder = recorder
	return r
}

func (r *Repository) WithResolvedConfigCache(cache *resolvedConfigCache) *Repository {
	if r == nil {
		return nil
	}
	r.configCache = cache
	if cache != nil {
		cache.Start(context.Background())
	}
	return r
}

func (r *Repository) WithDefaultUploadKey(key string) *Repository {
	if r == nil {
		return nil
	}
	r.defaultKey = strings.TrimSpace(key)
	return r
}

func normalizeTenantID(tenantID string) string {
	if tenantID == "" {
		return defaultTenantID
	}
	return tenantID
}

func setDefaultTenantID(tenantID string) {
	if tenantID == "" {
		defaultTenantID = "default"
		return
	}
	defaultTenantID = tenantID
}

func (r *Repository) EnsureProvider(ctx context.Context, provider *models.StorageProvider) error {
	if provider == nil {
		return errors.New("provider is nil")
	}
	provider.TenantID = normalizeTenantID(provider.TenantID)
	if provider.ID == uuid.Nil {
		provider.ID = uuid.New()
	}
	before, err := r.getProviderForAudit(ctx, provider.TenantID, provider.ProviderKey)
	if err != nil {
		return err
	}
	if err := r.encryptProviderSecrets(ctx, provider); err != nil {
		return err
	}
	if provider.Extra == nil {
		provider.Extra = models.MetaJSON{}
	}
	if before != nil {
		provider.ID = before.ID
		err = r.db.WithContext(ctx).Model(before).Updates(map[string]interface{}{
			"name": provider.Name, "driver": provider.Driver, "endpoint": provider.Endpoint,
			"region": provider.Region, "base_url": provider.BaseURL,
			"access_key_encrypted": provider.AccessKeyEncrypted, "secret_key_encrypted": provider.SecretKeyEncrypted,
			"extra": provider.Extra, "is_default": provider.IsDefault, "status": provider.Status,
			"updated_at": time.Now(), "deleted_at": nil,
		}).Error
	} else {
		err = r.db.WithContext(ctx).Create(provider).Error
	}
	if err != nil {
		return err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       buildUploadAuditAction("system.upload.provider", before != nil),
		ResourceType: "upload_provider",
		ResourceID:   provider.ProviderKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotProviderForAudit(before),
		After:        snapshotProviderForAudit(provider),
	})
	r.invalidateTenantUploadKeyCache(ctx, provider.TenantID)
	return nil
}

func (r *Repository) getProviderForAudit(ctx context.Context, tenantID, providerKey string) (*models.StorageProvider, error) {
	var item models.StorageProvider
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider_key = ? AND deleted_at IS NULL", tenantID, providerKey).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) EnsureBucket(ctx context.Context, bucket *models.StorageBucket) error {
	if bucket == nil {
		return errors.New("bucket is nil")
	}
	bucket.TenantID = normalizeTenantID(bucket.TenantID)
	if bucket.ID == uuid.Nil {
		bucket.ID = uuid.New()
	}
	before, err := r.getBucketForAudit(ctx, bucket.TenantID, bucket.BucketKey)
	if err != nil {
		return err
	}
	if bucket.Extra == nil {
		bucket.Extra = models.MetaJSON{}
	}
	if before != nil {
		bucket.ID = before.ID
		err = r.db.WithContext(ctx).Model(before).Updates(map[string]interface{}{
			"provider_id": bucket.ProviderID, "name": bucket.Name, "bucket_name": bucket.BucketName,
			"base_path": bucket.BasePath, "public_base_url": bucket.PublicBaseURL,
			"is_public": bucket.IsPublic, "status": bucket.Status, "extra": bucket.Extra,
			"updated_at": time.Now(), "deleted_at": nil,
		}).Error
	} else {
		err = r.db.WithContext(ctx).Create(bucket).Error
	}
	if err != nil {
		return err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       buildUploadAuditAction("system.upload.bucket", before != nil),
		ResourceType: "upload_bucket",
		ResourceID:   bucket.BucketKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotBucketForAudit(before),
		After:        snapshotBucketForAudit(bucket),
	})
	r.invalidateTenantUploadKeyCache(ctx, bucket.TenantID)
	return nil
}

func (r *Repository) GetProviderByKey(ctx context.Context, tenantID, providerKey string) (*models.StorageProvider, error) {
	tenantID = normalizeTenantID(tenantID)
	var provider models.StorageProvider
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider_key = ? AND deleted_at IS NULL", tenantID, providerKey).
		First(&provider).Error; err != nil {
		return nil, err
	}
	if err := r.decryptProviderSecrets(ctx, &provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *Repository) GetBucketByKey(ctx context.Context, tenantID, bucketKey string) (*models.StorageBucket, error) {
	tenantID = normalizeTenantID(tenantID)
	var bucket models.StorageBucket
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND bucket_key = ? AND deleted_at IS NULL", tenantID, bucketKey).
		First(&bucket).Error; err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (r *Repository) EnsureUploadKey(ctx context.Context, item *models.UploadKey) error {
	if item == nil {
		return errors.New("upload key is nil")
	}
	item.TenantID = normalizeTenantID(item.TenantID)
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	before, err := r.getUploadKeyForAudit(ctx, item.TenantID, item.Key)
	if err != nil {
		return err
	}
	if item.Meta == nil {
		item.Meta = models.MetaJSON{}
	}
	if item.AllowedMimeTypes == nil {
		item.AllowedMimeTypes = models.StringList{}
	}
	if before != nil {
		item.ID = before.ID
		err = r.db.WithContext(ctx).Model(before).Updates(map[string]interface{}{
			"bucket_id": item.BucketID, "name": item.Name, "path_template": item.PathTemplate,
			"default_rule_key": item.DefaultRuleKey, "max_size_bytes": item.MaxSizeBytes,
			"allowed_mime_types": item.AllowedMimeTypes, "visibility": item.Visibility,
			"status": item.Status, "meta": item.Meta, "updated_at": time.Now(), "deleted_at": nil,
		}).Error
	} else {
		err = r.db.WithContext(ctx).Create(item).Error
	}
	if err != nil {
		return err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       buildUploadAuditAction("system.upload.key", before != nil),
		ResourceType: "upload_key",
		ResourceID:   item.Key,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotUploadKeyForAudit(before),
		After:        snapshotUploadKeyForAudit(item),
	})
	storedItem, err := r.GetUploadKeyByKey(ctx, item.TenantID, item.Key)
	if err == nil {
		r.invalidateUploadConfigCache(ctx, item.TenantID, item.Key, storedItem.ID)
	} else {
		r.invalidateUploadConfigCache(ctx, item.TenantID, item.Key, uuid.Nil)
	}
	return nil
}

func (r *Repository) GetUploadKeyByKey(ctx context.Context, tenantID, key string) (*models.UploadKey, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadKey
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND key = ? AND deleted_at IS NULL", tenantID, key).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) EnsureUploadRule(ctx context.Context, item *models.UploadKeyRule) error {
	if item == nil {
		return errors.New("upload rule is nil")
	}
	item.TenantID = normalizeTenantID(item.TenantID)
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	before, err := r.getUploadRuleForAudit(ctx, item.TenantID, item.UploadKeyID, item.RuleKey)
	if err != nil {
		return err
	}
	if item.Meta == nil {
		item.Meta = models.MetaJSON{}
	}
	if item.AllowedMimeTypes == nil {
		item.AllowedMimeTypes = models.StringList{}
	}
	if item.ProcessPipeline == nil {
		item.ProcessPipeline = models.StringList{}
	}
	if before != nil {
		item.ID = before.ID
		err = r.db.WithContext(ctx).Model(before).Updates(map[string]interface{}{
			"name": item.Name, "sub_path": item.SubPath, "filename_strategy": item.FilenameStrategy,
			"max_size_bytes": item.MaxSizeBytes, "allowed_mime_types": item.AllowedMimeTypes,
			"process_pipeline": item.ProcessPipeline, "is_default": item.IsDefault,
			"status": item.Status, "meta": item.Meta, "updated_at": time.Now(), "deleted_at": nil,
		}).Error
	} else {
		err = r.db.WithContext(ctx).Create(item).Error
	}
	if err != nil {
		return err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       buildUploadAuditAction("system.upload.rule", before != nil),
		ResourceType: "upload_rule",
		ResourceID:   item.RuleKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotUploadRuleForAudit(before),
		After:        snapshotUploadRuleForAudit(item),
	})
	return nil
}

func (r *Repository) GetDefaultConfig(ctx context.Context, tenantID string) (*ResolvedConfig, error) {
	if defaultKey := strings.TrimSpace(r.defaultKey); defaultKey != "" {
		return r.GetResolvedConfigByKey(ctx, tenantID, defaultKey)
	}
	return r.loadDefaultConfig(ctx, tenantID)
}

func (r *Repository) GetResolvedConfigByKey(ctx context.Context, tenantID, key string) (*ResolvedConfig, error) {
	tenantID = normalizeTenantID(tenantID)
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("upload key is empty")
	}
	baseConfig, err := r.getCachedUploadConfig(ctx, tenantID, key, func(loadCtx context.Context) (*ResolvedConfig, error) {
		return r.loadResolvedConfigByKey(loadCtx, tenantID, key)
	})
	if err != nil {
		return nil, err
	}
	rule, err := r.getCachedUploadRule(ctx, baseConfig.UploadKey.ID, func(loadCtx context.Context) (*models.UploadKeyRule, error) {
		return r.loadResolvedRule(loadCtx, tenantID, baseConfig.UploadKey)
	})
	if err != nil {
		return nil, err
	}
	baseConfig.Rule = *rule
	return baseConfig, nil
}

func (r *Repository) CreateRecord(ctx context.Context, item *models.UploadRecord) error {
	if item == nil {
		return errors.New("upload record is nil")
	}
	item.TenantID = normalizeTenantID(item.TenantID)
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return err
	}
	// 直传链路收口：record 落库即代表「文件已存进对象存储 + 元数据写表」，
	// 必须留下一条审计行 —— 这是一切对外可见上传操作的真相源。
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.record.create",
		ResourceType: "upload_record",
		ResourceID:   item.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		After:        snapshotUploadRecordForAudit(item),
	})
	return nil
}

func (r *Repository) ListRecords(ctx context.Context, tenantID string, limit int) ([]models.UploadRecord, error) {
	tenantID = normalizeTenantID(tenantID)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var items []models.UploadRecord
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *Repository) CountRecords(ctx context.Context, tenantID string) (int64, error) {
	tenantID = normalizeTenantID(tenantID)
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.UploadRecord{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&total).Error
	return total, err
}

func (r *Repository) GetRecord(ctx context.Context, tenantID string, id uuid.UUID) (*models.UploadRecord, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadRecord
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) GetRecordByStorageKey(ctx context.Context, tenantID, storageKey string) (*models.UploadRecord, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadRecord
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND storage_key = ? AND deleted_at IS NULL", tenantID, strings.TrimSpace(storageKey)).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) DeleteRecord(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = normalizeTenantID(tenantID)
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.UploadRecord{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
			"status":     "deleted",
		}).Error
}

func (r *Repository) DeleteBucket(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = normalizeTenantID(tenantID)
	var bucket models.StorageBucket
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&bucket).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.StorageBucket{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
			"status":     "deleted",
		}).Error; err != nil {
		return err
	}
	after := bucket
	after.Status = "deleted"
	after.DeletedAt.Time = now
	after.DeletedAt.Valid = true
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.bucket.delete",
		ResourceType: "upload_bucket",
		ResourceID:   bucket.BucketKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotBucketForAudit(&bucket),
		After:        snapshotBucketForAudit(&after),
	})
	r.invalidateTenantUploadKeyCache(ctx, tenantID)
	return nil
}

func (r *Repository) DeleteUploadKey(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadKey
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.UploadKey{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
			"status":     "deleted",
		}).Error; err != nil {
		return err
	}
	after := item
	after.Status = "deleted"
	after.DeletedAt.Time = now
	after.DeletedAt.Valid = true
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.key.delete",
		ResourceType: "upload_key",
		ResourceID:   item.Key,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotUploadKeyForAudit(&item),
		After:        snapshotUploadKeyForAudit(&after),
	})
	r.invalidateUploadConfigCache(ctx, tenantID, item.Key, item.ID)
	return nil
}

func (r *Repository) DeleteUploadRule(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadKeyRule
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.UploadKeyRule{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
			"status":     "deleted",
		}).Error; err != nil {
		return err
	}
	after := item
	after.Status = "deleted"
	after.DeletedAt.Time = now
	after.DeletedAt.Valid = true
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.rule.delete",
		ResourceType: "upload_rule",
		ResourceID:   item.RuleKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotUploadRuleForAudit(&item),
		After:        snapshotUploadRuleForAudit(&after),
	})
	r.invalidateUploadRuleCache(ctx, item.UploadKeyID)
	return nil
}

func (r *Repository) loadDefaultConfig(ctx context.Context, tenantID string) (*ResolvedConfig, error) {
	tenantID = normalizeTenantID(tenantID)

	var provider models.StorageProvider
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_default = ? AND deleted_at IS NULL", tenantID, true).
		Order("updated_at DESC").
		First(&provider).Error; err != nil {
		return nil, err
	}
	if err := r.decryptProviderSecrets(ctx, &provider); err != nil {
		return nil, err
	}

	var bucket models.StorageBucket
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider_id = ? AND deleted_at IS NULL", tenantID, provider.ID).
		Order("updated_at DESC").
		First(&bucket).Error; err != nil {
		return nil, err
	}

	var uploadKey models.UploadKey
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND bucket_id = ? AND deleted_at IS NULL", tenantID, bucket.ID).
		Order("updated_at DESC").
		First(&uploadKey).Error; err != nil {
		return nil, err
	}

	rule, err := r.loadResolvedRule(ctx, tenantID, uploadKey)
	if err != nil {
		return nil, err
	}

	return &ResolvedConfig{
		Provider:  provider,
		Bucket:    bucket,
		UploadKey: uploadKey,
		Rule:      *rule,
	}, nil
}

func (r *Repository) loadResolvedConfigByKey(ctx context.Context, tenantID, key string) (*ResolvedConfig, error) {
	var uploadKey models.UploadKey
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND key = ? AND deleted_at IS NULL", tenantID, key).
		First(&uploadKey).Error; err != nil {
		return nil, err
	}

	var bucket models.StorageBucket
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, uploadKey.BucketID).
		First(&bucket).Error; err != nil {
		return nil, err
	}

	var provider models.StorageProvider
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, bucket.ProviderID).
		First(&provider).Error; err != nil {
		return nil, err
	}
	if err := r.decryptProviderSecrets(ctx, &provider); err != nil {
		return nil, err
	}

	return &ResolvedConfig{
		Provider:  provider,
		Bucket:    bucket,
		UploadKey: uploadKey,
	}, nil
}

func (r *Repository) loadResolvedRule(ctx context.Context, tenantID string, uploadKey models.UploadKey) (*models.UploadKeyRule, error) {
	var rule models.UploadKeyRule
	ruleQuery := r.db.WithContext(ctx).
		Where("tenant_id = ? AND upload_key_id = ? AND deleted_at IS NULL", tenantID, uploadKey.ID)
	if uploadKey.DefaultRuleKey != "" {
		ruleQuery = ruleQuery.Order(clause.Expr{SQL: "CASE WHEN rule_key = ? THEN 0 ELSE 1 END", Vars: []any{uploadKey.DefaultRuleKey}})
	}
	if err := ruleQuery.Order("is_default DESC, updated_at DESC").First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.UploadKeyRule{}, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *Repository) getCachedUploadConfig(ctx context.Context, tenantID, key string, loader func(context.Context) (*ResolvedConfig, error)) (*ResolvedConfig, error) {
	if r == nil || r.configCache == nil {
		return loader(ctx)
	}
	cacheKey := uploadKeyCacheKey(tenantID, key)
	var cached ResolvedConfig
	if r.configCache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}
	loaded, err := loader(ctx)
	if err != nil {
		return nil, err
	}
	r.configCache.Set(ctx, cacheKey, loaded)
	return loaded, nil
}

func (r *Repository) getCachedUploadRule(ctx context.Context, uploadKeyID uuid.UUID, loader func(context.Context) (*models.UploadKeyRule, error)) (*models.UploadKeyRule, error) {
	if r == nil || r.configCache == nil {
		return loader(ctx)
	}
	cacheKey := uploadRuleCacheKey(uploadKeyID)
	var cached models.UploadKeyRule
	if r.configCache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}
	loaded, err := loader(ctx)
	if err != nil {
		return nil, err
	}
	r.configCache.Set(ctx, cacheKey, loaded)
	return loaded, nil
}

func (r *Repository) invalidateTenantUploadKeyCache(ctx context.Context, tenantID string) {
	if r == nil || r.configCache == nil {
		return
	}
	r.configCache.Invalidate(ctx, nil, []string{uploadKeyCachePrefix(tenantID)})
}

func (r *Repository) invalidateUploadConfigCache(ctx context.Context, tenantID, key string, uploadKeyID uuid.UUID) {
	if r == nil || r.configCache == nil {
		return
	}
	keys := []string{uploadKeyCacheKey(tenantID, key)}
	if uploadKeyID != uuid.Nil {
		keys = append(keys, uploadRuleCacheKey(uploadKeyID))
	}
	r.configCache.Invalidate(ctx, keys, nil)
}

func (r *Repository) invalidateUploadRuleCache(ctx context.Context, uploadKeyID uuid.UUID) {
	if r == nil || r.configCache == nil || uploadKeyID == uuid.Nil {
		return
	}
	r.configCache.Invalidate(ctx, []string{uploadRuleCacheKey(uploadKeyID)}, nil)
}

type ResolvedConfig struct {
	Provider  models.StorageProvider
	Bucket    models.StorageBucket
	UploadKey models.UploadKey
	Rule      models.UploadKeyRule
}

type BucketConfig struct {
	Bucket   models.StorageBucket
	Provider models.StorageProvider
}

func (r *Repository) GetBucketDecrypted(ctx context.Context, tenantID string, bucketID uuid.UUID) (*BucketConfig, error) {
	tenantID = normalizeTenantID(tenantID)

	var bucket models.StorageBucket
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, bucketID).
		First(&bucket).Error; err != nil {
		return nil, err
	}

	var provider models.StorageProvider
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, bucket.ProviderID).
		First(&provider).Error; err != nil {
		return nil, err
	}
	if err := r.decryptProviderSecrets(ctx, &provider); err != nil {
		return nil, err
	}

	return &BucketConfig{
		Bucket:   bucket,
		Provider: provider,
	}, nil
}

func (r *Repository) GetBucketMasked(ctx context.Context, tenantID string, bucketID uuid.UUID) (*BucketConfig, error) {
	cfg, err := r.GetBucketDecrypted(ctx, tenantID, bucketID)
	if err != nil {
		return nil, err
	}
	cfg.Provider.AccessKeyEncrypted = maskSecretValue(cfg.Provider.AccessKeyEncrypted)
	cfg.Provider.SecretKeyEncrypted = maskSecretValue(cfg.Provider.SecretKeyEncrypted)
	return cfg, nil
}

func (r *Repository) encryptProviderSecrets(ctx context.Context, provider *models.StorageProvider) error {
	if provider == nil || r == nil || r.cipher == nil {
		return nil
	}
	encrypt := func(value string) (string, error) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || isEncryptedSecret(trimmed) {
			return value, nil
		}
		encrypted, err := r.cipher.Encrypt(ctx, trimmed)
		if err != nil {
			return "", err
		}
		return encrypted, nil
	}

	var err error
	provider.AccessKeyEncrypted, err = encrypt(provider.AccessKeyEncrypted)
	if err != nil {
		return fmt.Errorf("encrypt access key: %w", err)
	}
	provider.SecretKeyEncrypted, err = encrypt(provider.SecretKeyEncrypted)
	if err != nil {
		return fmt.Errorf("encrypt secret key: %w", err)
	}
	return nil
}

func (r *Repository) decryptProviderSecrets(ctx context.Context, provider *models.StorageProvider) error {
	if provider == nil || r == nil || r.cipher == nil {
		return nil
	}
	decrypt := func(value string) (string, error) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || !isEncryptedSecret(trimmed) {
			return value, nil
		}
		plaintext, err := r.cipher.Decrypt(ctx, trimmed)
		if err != nil {
			return "", err
		}
		return plaintext, nil
	}

	var err error
	provider.AccessKeyEncrypted, err = decrypt(provider.AccessKeyEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt access key: %w", err)
	}
	provider.SecretKeyEncrypted, err = decrypt(provider.SecretKeyEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt secret key: %w", err)
	}
	return nil
}

func isEncryptedSecret(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), secretCipherPrefix+":")
}

func maskSecretValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 4 {
		return "****"
	}
	return trimmed[:2] + strings.Repeat("*", len(trimmed)-4) + trimmed[len(trimmed)-2:]
}

func (r *Repository) recordAudit(ctx context.Context, event audit.Event) {
	if r == nil || r.auditRecorder == nil {
		return
	}
	r.auditRecorder.Record(ctx, event)
}

func buildUploadAuditAction(prefix string, updated bool) string {
	if updated {
		return prefix + ".update"
	}
	return prefix + ".create"
}

func (r *Repository) getBucketForAudit(ctx context.Context, tenantID, bucketKey string) (*models.StorageBucket, error) {
	var bucket models.StorageBucket
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND bucket_key = ? AND deleted_at IS NULL", tenantID, bucketKey).
		First(&bucket).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (r *Repository) getUploadKeyForAudit(ctx context.Context, tenantID, key string) (*models.UploadKey, error) {
	var item models.UploadKey
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND key = ? AND deleted_at IS NULL", tenantID, key).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) getUploadRuleForAudit(ctx context.Context, tenantID string, uploadKeyID uuid.UUID, ruleKey string) (*models.UploadKeyRule, error) {
	var item models.UploadKeyRule
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND upload_key_id = ? AND rule_key = ? AND deleted_at IS NULL", tenantID, uploadKeyID, ruleKey).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func snapshotBucketForAudit(bucket *models.StorageBucket) map[string]any {
	if bucket == nil {
		return nil
	}
	return map[string]any{
		"id":              bucket.ID.String(),
		"tenant_id":       bucket.TenantID,
		"provider_id":     bucket.ProviderID.String(),
		"bucket_key":      bucket.BucketKey,
		"name":            bucket.Name,
		"bucket_name":     bucket.BucketName,
		"base_path":       bucket.BasePath,
		"public_base_url": bucket.PublicBaseURL,
		"is_public":       bucket.IsPublic,
		"status":          bucket.Status,
	}
}

func snapshotUploadKeyForAudit(item *models.UploadKey) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":                 item.ID.String(),
		"tenant_id":          item.TenantID,
		"bucket_id":          item.BucketID.String(),
		"key":                item.Key,
		"name":               item.Name,
		"path_template":      item.PathTemplate,
		"default_rule_key":   item.DefaultRuleKey,
		"max_size_bytes":     item.MaxSizeBytes,
		"allowed_mime_types": item.AllowedMimeTypes,
		"visibility":         item.Visibility,
		"status":             item.Status,
	}
}

func snapshotUploadRecordForAudit(item *models.UploadRecord) map[string]any {
	if item == nil {
		return nil
	}
	uploaderID := ""
	if item.UploadedBy != nil {
		uploaderID = item.UploadedBy.String()
	}
	ruleID := ""
	if item.RuleID != nil {
		ruleID = item.RuleID.String()
	}
	source := ""
	if item.Meta != nil {
		if v, ok := item.Meta["source"].(string); ok {
			source = v
		}
	}
	return map[string]any{
		"id":                item.ID.String(),
		"tenant_id":         item.TenantID,
		"provider_id":       item.ProviderID.String(),
		"bucket_id":         item.BucketID.String(),
		"upload_key_id":     item.UploadKeyID.String(),
		"rule_id":           ruleID,
		"uploaded_by":       uploaderID,
		"original_filename": item.OriginalFilename,
		"stored_filename":   item.StoredFilename,
		"storage_key":       item.StorageKey,
		"url":               item.URL,
		"mime_type":         item.MimeType,
		"size":              item.Size,
		"checksum":          item.Checksum,
		"status":            item.Status,
		"source":            source,
	}
}

func snapshotUploadRuleForAudit(item *models.UploadKeyRule) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":                 item.ID.String(),
		"tenant_id":          item.TenantID,
		"upload_key_id":      item.UploadKeyID.String(),
		"rule_key":           item.RuleKey,
		"name":               item.Name,
		"sub_path":           item.SubPath,
		"filename_strategy":  item.FilenameStrategy,
		"max_size_bytes":     item.MaxSizeBytes,
		"allowed_mime_types": item.AllowedMimeTypes,
		"process_pipeline":   item.ProcessPipeline,
		"is_default":         item.IsDefault,
		"status":             item.Status,
	}
}

