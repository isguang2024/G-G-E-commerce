package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterEntry 公开注册入口配置（运行时唯一真相源）。
// 服务端按 host + path_prefix 命中入口，入口内联完整注册决策：
// 注册规则、验证码配置、注册后去向、默认绑定角色/功能包。
// 同一份后端可以挂多个入口（多 App、多渠道），通过 entry_code 唯一标识。
type RegisterEntry struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey         string         `gorm:"type:varchar(64);not null;index" json:"app_key"`
	EntryCode      string         `gorm:"type:varchar(64);not null;uniqueIndex" json:"entry_code"`
	Name           string         `gorm:"type:varchar(128);not null" json:"name"`
	Description    string         `gorm:"type:text;not null;default:''" json:"description"`
	Host           string         `gorm:"type:varchar(128);not null;default:''" json:"host"`
	PathPrefix     string         `gorm:"type:varchar(256);not null;default:''" json:"path_prefix"`
	RegisterSource string         `gorm:"type:varchar(32);not null;default:'self'" json:"register_source"`
	LoginPageKey   string         `gorm:"type:varchar(64);not null;default:''" json:"login_page_key"`
	Status         string         `gorm:"type:varchar(16);not null;default:'enabled'" json:"status"`

	// ── 注册规则（直接内联，不再从 policy 合并） ──
	AllowPublicRegister bool `gorm:"not null;default:false" json:"allow_public_register"`
	RequireInvite       bool `gorm:"not null;default:false" json:"require_invite"`
	RequireEmailVerify  bool `gorm:"not null;default:false" json:"require_email_verify"`
	RequireCaptcha      bool `gorm:"not null;default:false" json:"require_captcha"`
	AutoLogin           bool `gorm:"not null;default:true" json:"auto_login"`

	// ── 验证码配置 ──
	CaptchaProvider string `gorm:"type:varchar(32);not null;default:'none'" json:"captcha_provider"`
	CaptchaSiteKey  string `gorm:"type:varchar(256);not null;default:''" json:"captcha_site_key"`

	// ── 注册后去向（注册决策） ──
	TargetURL                 string `gorm:"type:varchar(1024);not null;default:''" json:"target_url"`
	TargetAppKey              string `gorm:"type:varchar(64);not null;default:''" json:"target_app_key"`
	TargetNavigationSpaceKey  string `gorm:"type:varchar(64);not null;default:''" json:"target_navigation_space_key"`
	TargetHomePath            string `gorm:"type:varchar(256);not null;default:''" json:"target_home_path"`
	WelcomeMessageTemplateKey string `gorm:"type:varchar(128);not null;default:''" json:"welcome_message_template_key"`

	// ── 注册决策：默认绑定的角色 code 和功能包 key ──
	RoleCodes          StringList `gorm:"type:jsonb;not null;default:'[]'" json:"role_codes"`
	FeaturePackageKeys StringList `gorm:"type:jsonb;not null;default:'[]'" json:"feature_package_keys"`

	// ── 系统标记 ──
	IsSystemReserved bool `gorm:"not null;default:false" json:"is_system_reserved"`

	SortOrder int            `gorm:"not null;default:0" json:"sort_order"`
	Remark    string         `gorm:"type:text;not null;default:''" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (RegisterEntry) TableName() string { return "register_entries" }
