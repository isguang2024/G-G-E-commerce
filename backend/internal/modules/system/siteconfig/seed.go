package siteconfig

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

// defaultSeedTenant 与 upload 模块保持一致，当前项目 tenant 固定为 default。
const defaultSeedTenant = "default"

// 默认集合编码。
const (
	SeedSetSiteBasic      = "site_basic"
	SeedSetSiteBranding   = "site_branding"
	SeedSetSiteFooter     = "site_footer"
	SeedSetPlatformGlobal = "platform_global"
)

// 默认配置 key。
const (
	SeedKeySiteName                = "site.name"
	SeedKeySiteDescription         = "site.description"
	SeedKeySiteLogo                = "site.logo"
	SeedKeySiteFavicon             = "site.favicon"
	SeedKeySiteCopyright           = "site.copyright"
	SeedKeyPlatformMaintenanceMode = "platform.maintenance_mode"
	SeedKeyPlatformRegistrationOpn = "platform.registration_open"
)

type defaultSetSpec struct {
	Code        string
	Name        string
	Description string
	SortOrder   int
}

type defaultConfigSpec struct {
	Key         string
	Label       string
	Description string
	ValueType   string
	Value       models.MetaJSON
	SortOrder   int
	Sets        []string
}

// SeedDefaults 写入站点配置的默认集合与全局配置项。幂等：存在则跳过，不覆盖用户已修改值。
func SeedDefaults(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	if db == nil {
		return errors.New("siteconfig seed: db is nil")
	}

	sets := []defaultSetSpec{
		{Code: SeedSetSiteBasic, Name: "站点基础", Description: "名称、描述等基础信息", SortOrder: 10},
		{Code: SeedSetSiteBranding, Name: "品牌形象", Description: "Logo / favicon 等视觉资源", SortOrder: 20},
		{Code: SeedSetSiteFooter, Name: "页脚信息", Description: "版权、备案、联系方式等", SortOrder: 30},
		{Code: SeedSetPlatformGlobal, Name: "平台开关", Description: "全平台生效的开关与策略", SortOrder: 40},
	}

	setIDByCode := make(map[string]uuid.UUID, len(sets))
	for _, spec := range sets {
		id, err := ensureSeedSet(ctx, db, spec)
		if err != nil {
			return err
		}
		setIDByCode[spec.Code] = id
	}

	configs := []defaultConfigSpec{
		{
			Key:         SeedKeySiteName,
			Label:       "站点名称",
			Description: "管理后台顶部与浏览器标题使用的名称",
			ValueType:   models.SiteConfigValueTypeString,
			Value:       models.MetaJSON{"value": "MaBen Admin"},
			SortOrder:   10,
			Sets:        []string{SeedSetSiteBasic},
		},
		{
			Key:         SeedKeySiteDescription,
			Label:       "站点描述",
			Description: "用于登录页 / SEO 描述",
			ValueType:   models.SiteConfigValueTypeString,
			Value:       models.MetaJSON{"value": "MaBen Admin 通用管理脚手架"},
			SortOrder:   20,
			Sets:        []string{SeedSetSiteBasic},
		},
		{
			Key:         SeedKeySiteLogo,
			Label:       "站点 Logo",
			Description: "侧边栏 / 顶栏显示的 logo，建议正方形 PNG",
			ValueType:   models.SiteConfigValueTypeImage,
			Value:       models.MetaJSON{"url": ""},
			SortOrder:   10,
			Sets:        []string{SeedSetSiteBranding},
		},
		{
			Key:         SeedKeySiteFavicon,
			Label:       "浏览器图标 (Favicon)",
			Description: "浏览器标签页图标，推荐 .ico / 32x32 PNG",
			ValueType:   models.SiteConfigValueTypeImage,
			Value:       models.MetaJSON{"url": ""},
			SortOrder:   20,
			Sets:        []string{SeedSetSiteBranding},
		},
		{
			Key:         SeedKeySiteCopyright,
			Label:       "版权文案",
			Description: "登录页 / 页脚显示的版权声明",
			ValueType:   models.SiteConfigValueTypeString,
			Value:       models.MetaJSON{"value": "© 2026 MaBen Admin. All rights reserved."},
			SortOrder:   10,
			Sets:        []string{SeedSetSiteFooter},
		},
		{
			Key:         SeedKeyPlatformMaintenanceMode,
			Label:       "维护模式",
			Description: "开启后，非管理员用户登录将被拒绝",
			ValueType:   models.SiteConfigValueTypeBool,
			Value:       models.MetaJSON{"value": false},
			SortOrder:   10,
			Sets:        []string{SeedSetPlatformGlobal},
		},
		{
			Key:         SeedKeyPlatformRegistrationOpn,
			Label:       "允许用户注册",
			Description: "控制前台注册入口是否可用",
			ValueType:   models.SiteConfigValueTypeBool,
			Value:       models.MetaJSON{"value": true},
			SortOrder:   20,
			Sets:        []string{SeedSetPlatformGlobal},
		},
	}

	for _, spec := range configs {
		if err := ensureSeedConfig(ctx, db, spec); err != nil {
			return err
		}
		for _, setCode := range spec.Sets {
			setID, ok := setIDByCode[setCode]
			if !ok {
				continue
			}
			if err := ensureSeedSetItem(ctx, db, setID, spec.Key); err != nil {
				return err
			}
		}
	}

	if logger != nil {
		logger.Info("Site config default seeds ensured",
			zap.Int("sets", len(sets)),
			zap.Int("configs", len(configs)),
		)
	}
	return nil
}

func ensureSeedSet(ctx context.Context, db *gorm.DB, spec defaultSetSpec) (uuid.UUID, error) {
	var existing models.SiteConfigSet
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND set_code = ?", defaultSeedTenant, spec.Code).
		First(&existing).Error
	if err == nil {
		// 仅补全内建标记与名称/描述的缺失字段，不覆盖用户自定义。
		updates := map[string]interface{}{"is_builtin": true}
		if existing.SetName == "" {
			updates["set_name"] = spec.Name
		}
		if existing.Description == "" {
			updates["description"] = spec.Description
		}
		if len(updates) == 0 {
			return existing.ID, nil
		}
		if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
			return uuid.Nil, err
		}
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	row := &models.SiteConfigSet{
		ID:          uuid.New(),
		TenantID:    defaultSeedTenant,
		SetCode:     spec.Code,
		SetName:     spec.Name,
		Description: spec.Description,
		SortOrder:   spec.SortOrder,
		IsBuiltin:   true,
		Status:      "normal",
	}
	if err := db.WithContext(ctx).Create(row).Error; err != nil {
		return uuid.Nil, err
	}
	return row.ID, nil
}

func ensureSeedConfig(ctx context.Context, db *gorm.DB, spec defaultConfigSpec) error {
	var existing models.SiteConfig
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND app_key = '' AND config_key = ?", defaultSeedTenant, spec.Key).
		First(&existing).Error
	if err == nil {
		// 已存在：仅补齐 label/description 等空字段，保留用户已设置的 value。
		updates := map[string]interface{}{"is_builtin": true}
		if existing.Label == "" {
			updates["label"] = spec.Label
		}
		if existing.Description == "" {
			updates["description"] = spec.Description
		}
		if existing.ValueType == "" {
			updates["value_type"] = spec.ValueType
		}
		if len(updates) == 1 && existing.IsBuiltin {
			return nil
		}
		return db.WithContext(ctx).Model(&existing).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	row := &models.SiteConfig{
		ID:          uuid.New(),
		TenantID:    defaultSeedTenant,
		AppKey:      "",
		ConfigKey:   spec.Key,
		ConfigValue: spec.Value,
		ValueType:   spec.ValueType,
		Label:       spec.Label,
		Description: spec.Description,
		SortOrder:   spec.SortOrder,
		IsBuiltin:   true,
		Status:      "normal",
	}
	if row.ConfigValue == nil {
		row.ConfigValue = models.MetaJSON{}
	}
	return db.WithContext(ctx).Create(row).Error
}

func ensureSeedSetItem(ctx context.Context, db *gorm.DB, setID uuid.UUID, key string) error {
	var existing models.SiteConfigSetItem
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND set_id = ? AND config_key = ?", defaultSeedTenant, setID, key).
		First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.WithContext(ctx).Create(&models.SiteConfigSetItem{
		ID:        uuid.New(),
		TenantID:  defaultSeedTenant,
		SetID:     setID,
		ConfigKey: key,
		SortOrder: 0,
	}).Error
}
