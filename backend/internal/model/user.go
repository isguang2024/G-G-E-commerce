package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email          string         `gorm:"type:varchar(255);uniqueIndex" json:"email"` // 邮箱改为可选
	Username       string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username"`
	PasswordHash   string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname       string         `gorm:"type:varchar(100)" json:"nickname"`
	AvatarURL      string         `gorm:"type:varchar(500)" json:"avatar_url"`
	Phone          string         `gorm:"type:varchar(20)" json:"phone"`
	SystemRemark   string         `gorm:"type:text" json:"system_remark"`
	Status         string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	IsSuperAdmin   bool           `gorm:"default:false" json:"is_super_admin"`
	LastLoginAt    *time.Time     `json:"last_login_at"`
	LastLoginIP    string         `gorm:"type:varchar(45)" json:"last_login_ip"`
	EmailVerifiedAt *time.Time    `json:"email_verified_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 关联关系
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
