package siteconfig

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

var (
	ErrConfigNotFound = errors.New("site config not found")
	ErrSetNotFound    = errors.New("site config set not found")
)

// Repository 参数管理仓储层。使用"先查后写"模式处理 Upsert，避免 GORM 的 ON CONFLICT 与部分唯一索引冲突。
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ---- SiteConfig ----

// ListConfigs 按 tenant + 作用域过滤。
// - global: 仅全局参数
// - app: 指定 APP 作用域参数
// - all: 全部作用域参数
func (r *Repository) ListConfigs(ctx context.Context, tenantID, scopeType, scopeKey string) ([]models.SiteConfig, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", normalizeTenantID(tenantID))
	switch normalizeListScopeType(scopeType) {
	case ScopeTypeAll:
		// all 不额外加条件
	case ScopeTypeApp:
		q = q.Where("app_key = ?", strings.TrimSpace(scopeKey))
	default:
		q = q.Where("app_key = ?", "")
	}
	var list []models.SiteConfig
	if err := q.Order("sort_order ASC, config_key ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetConfig 按 id 查询。
func (r *Repository) GetConfig(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfig, error) {
	var cfg models.SiteConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", normalizeTenantID(tenantID), id).First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrConfigNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ResolveByKeys 一次查询返回 app 级和全局两部分，由 service 层合并。
// 返回 map[app_key]map[config_key]*SiteConfig 形式，便于按作用域查找。
func (r *Repository) ResolveByKeys(ctx context.Context, tenantID, appKey string, keys []string) (appLevel, global map[string]*models.SiteConfig, err error) {
	if len(keys) == 0 {
		return map[string]*models.SiteConfig{}, map[string]*models.SiteConfig{}, nil
	}
	tenant := normalizeTenantID(tenantID)
	appKey = strings.TrimSpace(appKey)

	var list []models.SiteConfig
	q := r.db.WithContext(ctx).
		Where("tenant_id = ? AND config_key IN ? AND status = ?", tenant, keys, "normal").
		Where("app_key = '' OR app_key = ?", appKey).
		Find(&list)
	if q.Error != nil {
		return nil, nil, q.Error
	}

	appLevel = make(map[string]*models.SiteConfig)
	global = make(map[string]*models.SiteConfig)
	for i := range list {
		item := &list[i]
		if item.AppKey == "" {
			global[item.ConfigKey] = item
		} else if appKey != "" && item.AppKey == appKey {
			appLevel[item.ConfigKey] = item
		}
	}
	return appLevel, global, nil
}

// UpsertConfig 先查同作用域同 key 的记录，存在则更新，否则插入。
func (r *Repository) UpsertConfig(ctx context.Context, cfg *models.SiteConfig) error {
	if cfg == nil {
		return errors.New("site config is nil")
	}
	cfg.TenantID = normalizeTenantID(cfg.TenantID)
	cfg.AppKey = strings.TrimSpace(cfg.AppKey)
	cfg.ConfigKey = strings.TrimSpace(cfg.ConfigKey)
	if cfg.ConfigKey == "" {
		return errors.New("config_key is required")
	}
	if cfg.ValueType == "" {
		cfg.ValueType = models.SiteConfigValueTypeString
	}
	if cfg.FallbackPolicy == "" {
		cfg.FallbackPolicy = models.SiteConfigFallbackPolicyInherit
	}
	if cfg.Status == "" {
		cfg.Status = "normal"
	}
	if cfg.ConfigValue == nil {
		cfg.ConfigValue = models.MetaJSON{}
	}

	var existing models.SiteConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND app_key = ? AND config_key = ?", cfg.TenantID, cfg.AppKey, cfg.ConfigKey).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		cfg.ID = existing.ID
		return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"config_value":    cfg.ConfigValue,
			"value_type":      cfg.ValueType,
			"fallback_policy": cfg.FallbackPolicy,
			"label":           cfg.Label,
			"description":     cfg.Description,
			"sort_order":      cfg.SortOrder,
			"status":          cfg.Status,
			"updated_at":      time.Now(),
			"deleted_at":      nil,
		}).Error
	}
	if cfg.ID == uuid.Nil {
		cfg.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(cfg).Error
}

// DeleteConfig 软删除。返回被删除的记录，便于上层做缓存失效。
func (r *Repository) DeleteConfig(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfig, error) {
	cfg, err := r.GetConfig(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Delete(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

// ---- SiteConfigSet ----

func (r *Repository) ListSets(ctx context.Context, tenantID string) ([]models.SiteConfigSet, error) {
	var list []models.SiteConfigSet
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", normalizeTenantID(tenantID)).
		Order("sort_order ASC, set_code ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) GetSet(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfigSet, error) {
	var set models.SiteConfigSet
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", normalizeTenantID(tenantID), id).
		First(&set).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSetNotFound
	}
	if err != nil {
		return nil, err
	}
	return &set, nil
}

func (r *Repository) GetSetsByCodes(ctx context.Context, tenantID string, codes []string) ([]models.SiteConfigSet, error) {
	codes = compactStrings(codes)
	if len(codes) == 0 {
		return nil, nil
	}
	var list []models.SiteConfigSet
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND set_code IN ?", normalizeTenantID(tenantID), codes).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// UpsertSet 先查 set_code，存在则更新，否则插入。
func (r *Repository) UpsertSet(ctx context.Context, set *models.SiteConfigSet) error {
	if set == nil {
		return errors.New("site config set is nil")
	}
	set.TenantID = normalizeTenantID(set.TenantID)
	set.SetCode = strings.TrimSpace(set.SetCode)
	if set.SetCode == "" {
		return errors.New("set_code is required")
	}
	if set.Status == "" {
		set.Status = "normal"
	}

	var existing models.SiteConfigSet
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND set_code = ?", set.TenantID, set.SetCode).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		set.ID = existing.ID
		return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"set_name":    set.SetName,
			"description": set.Description,
			"sort_order":  set.SortOrder,
			"status":      set.Status,
			"updated_at":  time.Now(),
			"deleted_at":  nil,
		}).Error
	}
	if set.ID == uuid.Nil {
		set.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(set).Error
}

func (r *Repository) DeleteSet(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfigSet, error) {
	set, err := r.GetSet(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// items 通过外键 ON DELETE CASCADE，但软删时不级联，手动删除关联项
		if err := tx.Where("tenant_id = ? AND set_id = ?", set.TenantID, set.ID).Delete(&models.SiteConfigSetItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(set).Error
	})
	if err != nil {
		return nil, err
	}
	return set, nil
}

// ---- SiteConfigSetItem ----

func (r *Repository) ListItemsBySetIDs(ctx context.Context, tenantID string, setIDs []uuid.UUID) ([]models.SiteConfigSetItem, error) {
	if len(setIDs) == 0 {
		return nil, nil
	}
	var items []models.SiteConfigSetItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND set_id IN ?", normalizeTenantID(tenantID), setIDs).
		Order("sort_order ASC, config_key ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// ListItemsByConfigKey 反查某 key 属于哪些集合。
func (r *Repository) ListItemsByConfigKey(ctx context.Context, tenantID string, keys []string) ([]models.SiteConfigSetItem, error) {
	keys = compactStrings(keys)
	if len(keys) == 0 {
		return nil, nil
	}
	var items []models.SiteConfigSetItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND config_key IN ?", normalizeTenantID(tenantID), keys).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// ReplaceSetItems 整体替换某集合的 keys 列表。
func (r *Repository) ReplaceSetItems(ctx context.Context, tenantID string, setID uuid.UUID, keys []string) error {
	tenant := normalizeTenantID(tenantID)
	keys = compactStrings(keys)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND set_id = ?", tenant, setID).Delete(&models.SiteConfigSetItem{}).Error; err != nil {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		items := make([]models.SiteConfigSetItem, 0, len(keys))
		for i, k := range keys {
			items = append(items, models.SiteConfigSetItem{
				ID:        uuid.New(),
				TenantID:  tenant,
				SetID:     setID,
				ConfigKey: k,
				SortOrder: i,
			})
		}
		return tx.Create(&items).Error
	})
}
