package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DefaultAppKey  = "platform-admin"
	DefaultAppName = "平台管理后台"

	AppScopeShared = "shared"
	AppScopeApp    = "app"
)

type App struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey           string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"app_key"`
	Name             string         `gorm:"type:varchar(150);not null" json:"name"`
	Description      string         `gorm:"type:text;not null;default:''" json:"description"`
	SpaceMode        string         `gorm:"type:varchar(20);not null;default:'single'" json:"space_mode"`
	DefaultSpaceKey  string         `gorm:"type:varchar(100);not null;default:'default'" json:"default_space_key"`
	AuthMode         string         `gorm:"type:varchar(30);not null;default:'inherit_host'" json:"auth_mode"`
	FrontendEntryURL string         `gorm:"type:varchar(500);not null;default:''" json:"frontend_entry_url"`
	BackendEntryURL  string         `gorm:"type:varchar(500);not null;default:''" json:"backend_entry_url"`
	HealthCheckURL   string         `gorm:"type:varchar(500);not null;default:''" json:"health_check_url"`
	Capabilities     MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"capabilities"`
	Status           string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	IsDefault        bool           `gorm:"not null;default:false" json:"is_default"`
	Meta             MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (App) TableName() string {
	return "apps"
}

// 入口解析匹配类型。
const (
	EntryMatchHostExact   = "host_exact"
	EntryMatchHostSuffix  = "host_suffix"
	EntryMatchPathPrefix  = "path_prefix"
	EntryMatchHostAndPath = "host_and_path"
)

// AppHostBinding 兼容旧名（数据库表 app_host_bindings），承载 Level 1 的 APP 入口解析绑定。
type AppHostBinding struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey          string         `gorm:"type:varchar(100);not null;index" json:"app_key"`
	MatchType       string         `gorm:"type:varchar(30);not null;default:'host_exact'" json:"match_type"`
	Host            string         `gorm:"type:varchar(255);not null;default:''" json:"host"`
	PathPattern     string         `gorm:"type:varchar(255);not null;default:''" json:"path_pattern"`
	Priority        int            `gorm:"not null;default:0" json:"priority"`
	Description     string         `gorm:"type:text;not null;default:''" json:"description"`
	IsPrimary       bool           `gorm:"not null;default:false" json:"is_primary"`
	DefaultSpaceKey string         `gorm:"type:varchar(100);not null;default:'default'" json:"default_space_key"`
	Status          string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta            MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AppHostBinding) TableName() string {
	return "app_host_bindings"
}

// MenuSpaceEntryBinding Level 2 菜单空间入口解析绑定。
// 注意：与 MenuSpaceHostBinding（SSO/Cookie 配置）不同，这是路由匹配规则。
type MenuSpaceEntryBinding struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey      string         `gorm:"type:varchar(100);not null;index" json:"app_key"`
	SpaceKey    string         `gorm:"type:varchar(100);not null;index" json:"space_key"`
	MatchType   string         `gorm:"type:varchar(30);not null;default:'host_exact'" json:"match_type"`
	Host        string         `gorm:"type:varchar(255);not null;default:''" json:"host"`
	PathPattern string         `gorm:"type:varchar(255);not null;default:''" json:"path_pattern"`
	Priority    int            `gorm:"not null;default:0" json:"priority"`
	IsPrimary   bool           `gorm:"not null;default:false" json:"is_primary"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta        MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuSpaceEntryBinding) TableName() string {
	return "menu_space_entry_bindings"
}

func DefaultPlatformAdminCapabilities() MetaJSON {
	return MetaJSON{
		"auth": MetaJSON{
			"is_auth_center": false,
			"login_strategy": "centralized_login",
			"session_mode":   "token_exchange",
			"sso_mode":       "participate",
			"login_ui_mode":  "auth_center_ui",
			"login_page_key": "default",
		},
		"routing": MetaJSON{
			"entry_mode":              "inherit_host",
			"route_prefix":            "/",
			"supports_public_runtime": false,
		},
		"runtime": MetaJSON{
			"kind":                    "local",
			"supports_dynamic_routes": true,
			"supports_worktab":        true,
		},
		"navigation": MetaJSON{
			"supports_multi_space":  true,
			"default_landing_mode":  "menu_space",
			"supports_space_badges": true,
		},
		"integration": MetaJSON{
			"supports_app_switch":        true,
			"supports_broadcast_channel": true,
		},
	}
}

func DefaultAccountPortalCapabilities() MetaJSON {
	return MetaJSON{
		"auth": MetaJSON{
			"is_auth_center": true,
			"login_strategy": "local",
			"session_mode":   "first_party",
			"sso_mode":       "participate",
			"login_ui_mode":  "auth_center_ui",
			"login_page_key": "default",
		},
		"routing": MetaJSON{
			"entry_mode":              "path_prefix",
			"route_prefix":            "/account",
			"supports_public_runtime": true,
		},
		"runtime": MetaJSON{
			"kind":                    "local",
			"supports_dynamic_routes": true,
			"supports_worktab":        false,
		},
		"navigation": MetaJSON{
			"supports_multi_space": false,
			"default_landing_mode": "menu_space",
		},
		"integration": MetaJSON{
			"supports_app_switch":        true,
			"supports_broadcast_channel": false,
		},
	}
}
