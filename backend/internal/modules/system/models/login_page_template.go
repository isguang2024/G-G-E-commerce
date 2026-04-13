package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoginPageTemplate 登录/注册/找回密码统一认证页模板。
// tenant_id 预留多租户隔离，当前默认值为 default，所有查询必须显式过滤。
type LoginPageTemplate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	TemplateKey string         `gorm:"type:varchar(64);not null" json:"template_key"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Scene       string         `gorm:"type:varchar(32);not null;default:'auth_family'" json:"scene"`
	AppScope    string         `gorm:"type:varchar(32);not null;default:'shared'" json:"app_scope"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	IsDefault   bool           `gorm:"not null;default:false" json:"is_default"`
	Config      MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"config"`
	Meta        MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (LoginPageTemplate) TableName() string { return "login_page_templates" }
