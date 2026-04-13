package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterEntry 公开注册入口配置：服务端按 host + path_prefix 命中入口，
// 决定该次注册命中哪个 register source / policy。同一份后端可以挂多个入口
// （多 App、多渠道），通过 entry_code 唯一标识。
type RegisterEntry struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey              string         `gorm:"type:varchar(64);not null;index" json:"app_key"`
	EntryCode           string         `gorm:"type:varchar(64);not null;uniqueIndex" json:"entry_code"`
	Name                string         `gorm:"type:varchar(128);not null" json:"name"`
	Host                string         `gorm:"type:varchar(128);not null;default:''" json:"host"`
	PathPrefix          string         `gorm:"type:varchar(256);not null;default:''" json:"path_prefix"`
	RegisterSource      string         `gorm:"type:varchar(32);not null;default:'self'" json:"register_source"`
	PolicyCode          string         `gorm:"type:varchar(64);not null" json:"policy_code"`
	LoginPageKey        string         `gorm:"type:varchar(64);not null;default:''" json:"login_page_key"`
	Status              string         `gorm:"type:varchar(16);not null;default:'enabled'" json:"status"`
	AllowPublicRegister *bool          `json:"allow_public_register,omitempty"`
	RequireInvite       *bool          `json:"require_invite,omitempty"`
	RequireEmailVerify  *bool          `json:"require_email_verify,omitempty"`
	RequireCaptcha      *bool          `json:"require_captcha,omitempty"`
	AutoLogin           *bool          `json:"auto_login,omitempty"`
	SortOrder           int            `gorm:"not null;default:0" json:"sort_order"`
	Remark              string         `gorm:"type:text;not null;default:''" json:"remark"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (RegisterEntry) TableName() string { return "register_entries" }

// RegisterPolicy 注册策略：决定注册成功后用户进入哪个 App / MenuSpace / 首页，
// 以及绑定哪些默认功能包和角色。target_app_key 与入口 app_key 解耦，便于未来
// 把"入口归属 App"和"业务承载 App"拆开。
type RegisterPolicy struct {
	ID                        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey                    string    `gorm:"type:varchar(64);not null;index" json:"app_key"`
	PolicyCode                string    `gorm:"type:varchar(64);not null;uniqueIndex" json:"policy_code"`
	Name                      string    `gorm:"type:varchar(128);not null" json:"name"`
	Description               string    `gorm:"type:text;not null;default:''" json:"description"`
	TargetAppKey              string    `gorm:"type:varchar(64);not null" json:"target_app_key"`
	TargetNavigationSpaceKey  string    `gorm:"type:varchar(64);not null" json:"target_navigation_space_key"`
	TargetHomePath            string    `gorm:"type:varchar(256);not null;default:''" json:"target_home_path"`
	DefaultWorkspaceType      string    `gorm:"type:varchar(32);not null;default:'personal'" json:"default_workspace_type"`
	Status                    string    `gorm:"type:varchar(16);not null;default:'enabled'" json:"status"`
	WelcomeMessageTemplateKey string    `gorm:"type:varchar(128);not null;default:''" json:"welcome_message_template_key"`
	AllowPublicRegister       bool      `gorm:"not null;default:false" json:"allow_public_register"`
	RequireInvite             bool      `gorm:"not null;default:false" json:"require_invite"`
	RequireEmailVerify        bool      `gorm:"not null;default:false" json:"require_email_verify"`
	RequireCaptcha            bool      `gorm:"not null;default:false" json:"require_captcha"`
	AutoLogin                 bool      `gorm:"not null;default:true" json:"auto_login"`
	// 人机验证提供商：none | recaptcha | hcaptcha | turnstile
	CaptchaProvider string `gorm:"type:varchar(32);not null;default:'none'" json:"captcha_provider"`
	// 对应提供商的公开 site_key，前端渲染 captcha widget 使用
	CaptchaSiteKey string         `gorm:"type:varchar(256);not null;default:''" json:"captcha_site_key"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (RegisterPolicy) TableName() string { return "register_policies" }

// RegisterPolicyFeaturePackage 注册策略默认绑定的功能包。
type RegisterPolicyFeaturePackage struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PolicyCode     string    `gorm:"type:varchar(64);not null;index" json:"policy_code"`
	PackageID      uuid.UUID `gorm:"type:uuid;not null;index" json:"package_id"`
	WorkspaceScope string    `gorm:"type:varchar(32);not null;default:'personal'" json:"workspace_scope"`
	SortOrder      int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
}

func (RegisterPolicyFeaturePackage) TableName() string { return "register_policy_feature_packages" }

// RegisterPolicyRole 注册策略默认绑定的角色。
type RegisterPolicyRole struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PolicyCode     string    `gorm:"type:varchar(64);not null;index" json:"policy_code"`
	RoleID         uuid.UUID `gorm:"type:uuid;not null;index" json:"role_id"`
	WorkspaceScope string    `gorm:"type:varchar(32);not null;default:'personal'" json:"workspace_scope"`
	SortOrder      int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
}

func (RegisterPolicyRole) TableName() string { return "register_policy_roles" }
