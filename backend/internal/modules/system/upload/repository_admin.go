package upload

// repository_admin.go 提供上传配置中心管理面所需的 List/Get(by id)/Update/Delete 方法。
// 与 EnsureProvider/EnsureBucket/EnsureUploadKey/EnsureUploadRule（按 key upsert）互补：
// 管理面以 ID 为操作主键，避免 key 唯一约束在改名时冲突。

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/models"
)

// ---------- StorageProvider ----------

func (r *Repository) ListProviders(ctx context.Context, tenantID string) ([]models.StorageProvider, error) {
	tenantID = normalizeTenantID(tenantID)
	var items []models.StorageProvider
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("is_default DESC, updated_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetProviderByID(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageProvider, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.StorageProvider
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetProviderDecryptedByID returns provider with secrets decrypted (for connectivity tests).
func (r *Repository) GetProviderDecryptedByID(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageProvider, error) {
	item, err := r.GetProviderByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if err := r.decryptProviderSecrets(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// ProviderUpdate carries the writable fields for UpdateProviderByID.
// Any plain-text access_key / secret_key passed here is encrypted before writing.
type ProviderUpdate struct {
	ProviderKey string
	Name        string
	Driver      string
	Endpoint    string
	Region      string
	BaseURL     string
	AccessKey   string // plain-text; empty means keep current value
	SecretKey   string // plain-text; empty means keep current value
	IsDefault   bool
	Status      string
	Extra       models.MetaJSON
}

func (r *Repository) UpdateProviderByID(ctx context.Context, tenantID string, id uuid.UUID, update ProviderUpdate) (*models.StorageProvider, error) {
	tenantID = normalizeTenantID(tenantID)
	current, err := r.GetProviderByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	before := *current
	before.AccessKeyEncrypted = maskSecretValue(before.AccessKeyEncrypted)
	before.SecretKeyEncrypted = maskSecretValue(before.SecretKeyEncrypted)

	now := time.Now()
	updates := map[string]any{
		"provider_key": strings.TrimSpace(update.ProviderKey),
		"name":         strings.TrimSpace(update.Name),
		"driver":       strings.TrimSpace(update.Driver),
		"endpoint":     update.Endpoint,
		"region":       update.Region,
		"base_url":     update.BaseURL,
		"is_default":   update.IsDefault,
		"status":       firstNonEmpty(update.Status, models.UploadProviderStatusReady),
		"extra":        update.Extra,
		"updated_at":   now,
	}
	if strings.TrimSpace(update.AccessKey) != "" {
		encrypted, err := r.encryptSecretValue(ctx, update.AccessKey)
		if err != nil {
			return nil, err
		}
		updates["access_key_encrypted"] = encrypted
	}
	if strings.TrimSpace(update.SecretKey) != "" {
		encrypted, err := r.encryptSecretValue(ctx, update.SecretKey)
		if err != nil {
			return nil, err
		}
		updates["secret_key_encrypted"] = encrypted
	}
	if err := r.db.WithContext(ctx).
		Model(&models.StorageProvider{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	if update.IsDefault {
		if err := r.db.WithContext(ctx).
			Model(&models.StorageProvider{}).
			Where("tenant_id = ? AND id <> ? AND is_default = ? AND deleted_at IS NULL", tenantID, id, true).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}
	updated, err := r.GetProviderByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	after := *updated
	after.AccessKeyEncrypted = maskSecretValue(after.AccessKeyEncrypted)
	after.SecretKeyEncrypted = maskSecretValue(after.SecretKeyEncrypted)

	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.provider.update",
		ResourceType: "upload_provider",
		ResourceID:   updated.ProviderKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotProviderForAudit(&before),
		After:        snapshotProviderForAudit(&after),
	})
	r.invalidateTenantUploadKeyCache(ctx, tenantID)
	return updated, nil
}

func (r *Repository) DeleteProvider(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = normalizeTenantID(tenantID)
	current, err := r.GetProviderByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&models.StorageProvider{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
			"status":     "deleted",
			"is_default": false,
		}).Error; err != nil {
		return err
	}
	after := *current
	after.Status = "deleted"
	after.IsDefault = false
	after.DeletedAt.Time = now
	after.DeletedAt.Valid = true
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.provider.delete",
		ResourceType: "upload_provider",
		ResourceID:   current.ProviderKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotProviderForAudit(current),
		After:        snapshotProviderForAudit(&after),
	})
	r.invalidateTenantUploadKeyCache(ctx, tenantID)
	return nil
}

// ---------- StorageBucket ----------

func (r *Repository) ListBuckets(ctx context.Context, tenantID string, providerID *uuid.UUID) ([]models.StorageBucket, error) {
	tenantID = normalizeTenantID(tenantID)
	tx := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if providerID != nil && *providerID != uuid.Nil {
		tx = tx.Where("provider_id = ?", *providerID)
	}
	var items []models.StorageBucket
	if err := tx.Order("updated_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetBucketByID(ctx context.Context, tenantID string, id uuid.UUID) (*models.StorageBucket, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.StorageBucket
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

type BucketUpdate struct {
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

func (r *Repository) UpdateBucketByID(ctx context.Context, tenantID string, id uuid.UUID, update BucketUpdate) (*models.StorageBucket, error) {
	tenantID = normalizeTenantID(tenantID)
	current, err := r.GetBucketByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	before := *current
	now := time.Now()
	updates := map[string]any{
		"provider_id":     update.ProviderID,
		"bucket_key":      strings.TrimSpace(update.BucketKey),
		"name":            strings.TrimSpace(update.Name),
		"bucket_name":     strings.TrimSpace(update.BucketName),
		"base_path":       update.BasePath,
		"public_base_url": update.PublicBaseURL,
		"is_public":       update.IsPublic,
		"status":          firstNonEmpty(update.Status, models.UploadProviderStatusReady),
		"extra":           update.Extra,
		"updated_at":      now,
	}
	if err := r.db.WithContext(ctx).
		Model(&models.StorageBucket{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	updated, err := r.GetBucketByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.bucket.update",
		ResourceType: "upload_bucket",
		ResourceID:   updated.BucketKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotBucketForAudit(&before),
		After:        snapshotBucketForAudit(updated),
	})
	r.invalidateTenantUploadKeyCache(ctx, tenantID)
	return updated, nil
}

// ---------- UploadKey ----------

func (r *Repository) ListUploadKeys(ctx context.Context, tenantID string, bucketID *uuid.UUID) ([]models.UploadKey, error) {
	tenantID = normalizeTenantID(tenantID)
	tx := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if bucketID != nil && *bucketID != uuid.Nil {
		tx = tx.Where("bucket_id = ?", *bucketID)
	}
	var items []models.UploadKey
	if err := tx.Order("updated_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetUploadKeyByID(ctx context.Context, tenantID string, id uuid.UUID) (*models.UploadKey, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadKey
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListRulesByUploadKey(ctx context.Context, tenantID string, uploadKeyID uuid.UUID) ([]models.UploadKeyRule, error) {
	tenantID = normalizeTenantID(tenantID)
	var items []models.UploadKeyRule
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND upload_key_id = ? AND deleted_at IS NULL", tenantID, uploadKeyID).
		Order("is_default DESC, updated_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

type UploadKeyUpdate struct {
	BucketID         uuid.UUID
	Key              string
	Name             string
	PathTemplate     string
	DefaultRuleKey   string
	MaxSizeBytes     int64
	AllowedMimeTypes models.StringList
	Visibility       string
	Status           string
	Meta             models.MetaJSON
}

func (r *Repository) UpdateUploadKeyByID(ctx context.Context, tenantID string, id uuid.UUID, update UploadKeyUpdate) (*models.UploadKey, error) {
	tenantID = normalizeTenantID(tenantID)
	current, err := r.GetUploadKeyByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	before := *current
	now := time.Now()
	allowed := update.AllowedMimeTypes
	if allowed == nil {
		allowed = models.StringList{}
	}
	updates := map[string]any{
		"bucket_id":          update.BucketID,
		"key":                strings.TrimSpace(update.Key),
		"name":               strings.TrimSpace(update.Name),
		"path_template":      update.PathTemplate,
		"default_rule_key":   strings.TrimSpace(update.DefaultRuleKey),
		"max_size_bytes":     update.MaxSizeBytes,
		"allowed_mime_types": allowed,
		"visibility":         firstNonEmpty(update.Visibility, "public"),
		"status":             firstNonEmpty(update.Status, models.UploadProviderStatusReady),
		"meta":               update.Meta,
		"updated_at":         now,
	}
	if err := r.db.WithContext(ctx).
		Model(&models.UploadKey{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	updated, err := r.GetUploadKeyByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.key.update",
		ResourceType: "upload_key",
		ResourceID:   updated.Key,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotUploadKeyForAudit(&before),
		After:        snapshotUploadKeyForAudit(updated),
	})
	r.invalidateUploadConfigCache(ctx, tenantID, before.Key, before.ID)
	if before.Key != updated.Key {
		r.invalidateUploadConfigCache(ctx, tenantID, updated.Key, updated.ID)
	}
	return updated, nil
}

// ---------- UploadKeyRule ----------

func (r *Repository) GetUploadRuleByID(ctx context.Context, tenantID string, id uuid.UUID) (*models.UploadKeyRule, error) {
	tenantID = normalizeTenantID(tenantID)
	var item models.UploadKeyRule
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

type UploadRuleUpdate struct {
	RuleKey          string
	Name             string
	SubPath          string
	FilenameStrategy string
	MaxSizeBytes     int64
	AllowedMimeTypes models.StringList
	ProcessPipeline  models.StringList
	IsDefault        bool
	Status           string
	Meta             models.MetaJSON
}

func (r *Repository) UpdateUploadRuleByID(ctx context.Context, tenantID string, id uuid.UUID, update UploadRuleUpdate) (*models.UploadKeyRule, error) {
	tenantID = normalizeTenantID(tenantID)
	current, err := r.GetUploadRuleByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	before := *current
	now := time.Now()
	allowed := update.AllowedMimeTypes
	if allowed == nil {
		allowed = models.StringList{}
	}
	pipeline := update.ProcessPipeline
	if pipeline == nil {
		pipeline = models.StringList{}
	}
	updates := map[string]any{
		"rule_key":           strings.TrimSpace(update.RuleKey),
		"name":               strings.TrimSpace(update.Name),
		"sub_path":           update.SubPath,
		"filename_strategy":  firstNonEmpty(update.FilenameStrategy, "uuid"),
		"max_size_bytes":     update.MaxSizeBytes,
		"allowed_mime_types": allowed,
		"process_pipeline":   pipeline,
		"is_default":         update.IsDefault,
		"status":             firstNonEmpty(update.Status, models.UploadProviderStatusReady),
		"meta":               update.Meta,
		"updated_at":         now,
	}
	if err := r.db.WithContext(ctx).
		Model(&models.UploadKeyRule{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	if update.IsDefault {
		if err := r.db.WithContext(ctx).
			Model(&models.UploadKeyRule{}).
			Where("tenant_id = ? AND upload_key_id = ? AND id <> ? AND is_default = ? AND deleted_at IS NULL", tenantID, current.UploadKeyID, id, true).
			Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}
	updated, err := r.GetUploadRuleByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	r.recordAudit(ctx, audit.Event{
		Action:       "system.upload.rule.update",
		ResourceType: "upload_rule",
		ResourceID:   updated.RuleKey,
		Outcome:      audit.OutcomeSuccess,
		Before:       snapshotUploadRuleForAudit(&before),
		After:        snapshotUploadRuleForAudit(updated),
	})
	r.invalidateUploadRuleCache(ctx, current.UploadKeyID)
	return updated, nil
}

// ---------- helpers ----------

func (r *Repository) encryptSecretValue(ctx context.Context, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if isEncryptedSecret(trimmed) {
		return trimmed, nil
	}
	if r == nil || r.cipher == nil {
		// no cipher configured; persist plain text but mark with an obvious sentinel
		return "", errors.New("upload secret cipher is not configured; refusing to store plaintext credentials")
	}
	return r.cipher.Encrypt(ctx, trimmed)
}

func snapshotProviderForAudit(item *models.StorageProvider) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":           item.ID.String(),
		"tenant_id":    item.TenantID,
		"provider_key": item.ProviderKey,
		"name":         item.Name,
		"driver":       item.Driver,
		"endpoint":     item.Endpoint,
		"region":       item.Region,
		"base_url":     item.BaseURL,
		"is_default":   item.IsDefault,
		"status":       item.Status,
	}
}

// ConvertGormNotFound is a small helper for service-layer mapping.
func ConvertGormNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}
	return err
}

