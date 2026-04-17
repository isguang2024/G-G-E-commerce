package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 配置值类型
const (
	SiteConfigValueTypeString = "string"
	SiteConfigValueTypeNumber = "number"
	SiteConfigValueTypeBool   = "bool"
	SiteConfigValueTypeImage  = "image"
	SiteConfigValueTypeJSON   = "json"
	SiteConfigValueTypeSVG    = "svg" // 内联 SVG 标记文本，存储为 {"value": "<svg>...</svg>"}
)

// 回退策略。
const (
	SiteConfigFallbackPolicyInherit = "inherit"
	SiteConfigFallbackPolicyStrict  = "strict"
)

// SiteConfig 对应 site_configs 表。
// 当前存储层仍以 app_key 承载参数作用域：
// - app_key 为空表示全局作用域
// - app_key 非空表示某个 APP 作用域
// 回退策略：
// - inherit：当前作用域未命中时，允许回退到全局作用域
// - strict：当前作用域未命中时，不回退到全局作用域
type SiteConfig struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	AppKey         string         `gorm:"type:varchar(100);not null;default:''" json:"app_key"`
	ConfigKey      string         `gorm:"type:varchar(150);not null" json:"config_key"`
	ConfigValue    MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"config_value"`
	ValueType      string         `gorm:"type:varchar(32);not null;default:'string'" json:"value_type"`
	FallbackPolicy string         `gorm:"type:varchar(20);not null;default:'inherit'" json:"fallback_policy"`
	Label          string         `gorm:"type:varchar(200);not null;default:''" json:"label"`
	Description    string         `gorm:"type:varchar(500);not null;default:''" json:"description"`
	SortOrder      int            `gorm:"not null;default:0" json:"sort_order"`
	IsBuiltin      bool           `gorm:"not null;default:false" json:"is_builtin"`
	Status         string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SiteConfig) TableName() string { return "site_configs" }

// IsGlobal 返回是否为全局作用域参数（app_key 为空）。
func (m SiteConfig) IsGlobal() bool { return m.AppKey == "" }

// FallbackPolicyOrDefault 返回标准化后的回退策略。
func (m SiteConfig) FallbackPolicyOrDefault() string {
	switch m.FallbackPolicy {
	case SiteConfigFallbackPolicyStrict:
		return SiteConfigFallbackPolicyStrict
	default:
		return SiteConfigFallbackPolicyInherit
	}
}

// SiteConfigSet 参数集合（纯分组元数据，跨作用域共享）。
type SiteConfigSet struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	SetCode     string         `gorm:"type:varchar(100);not null" json:"set_code"`
	SetName     string         `gorm:"type:varchar(200);not null" json:"set_name"`
	Description string         `gorm:"type:varchar(500);not null;default:''" json:"description"`
	SortOrder   int            `gorm:"not null;default:0" json:"sort_order"`
	IsBuiltin   bool           `gorm:"not null;default:false" json:"is_builtin"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SiteConfigSet) TableName() string { return "site_config_sets" }

// SiteConfigSetItem 集合↔参数 key 多对多关联。config_key 是字符串（不是外键），
// 因为同一个 key 可能同时存在于全局和多个 APP 作用域参数中。
type SiteConfigSetItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string    `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	SetID     uuid.UUID `gorm:"type:uuid;not null;index" json:"set_id"`
	ConfigKey string    `gorm:"type:varchar(150);not null;index" json:"config_key"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (SiteConfigSetItem) TableName() string { return "site_config_set_items" }
