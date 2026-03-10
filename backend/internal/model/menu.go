package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetaJSON 菜单 meta 信息（标题、图标、角色等）
type MetaJSON map[string]interface{}

func (m MetaJSON) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *MetaJSON) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, m)
}

// Menu 菜单模型（树形，与前端路由结构对应）
type Menu struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ParentID  *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id"`                    // 父菜单 ID，空为顶级
	Path      string         `gorm:"type:varchar(255)" json:"path"`                      // 路由路径
	Name      string         `gorm:"type:varchar(100);index" json:"name"`                 // 路由 name
	Component string         `gorm:"type:varchar(255)" json:"component"`                // 组件路径，如 /system/user
	Title     string         `gorm:"type:varchar(100)" json:"title"`                     // 菜单标题
	Icon      string         `gorm:"type:varchar(100)" json:"icon"`                     // 图标
	SortOrder int            `gorm:"default:0" json:"sort_order"`                       // 排序
	Meta      MetaJSON       `gorm:"type:jsonb" json:"meta,omitempty"`                   // 扩展 meta（roles、isHide 等）
	Hidden    bool           `gorm:"default:false" json:"hidden"`                       // 是否隐藏
	IsSystem  bool           `gorm:"-" json:"is_system"`                                // 系统默认菜单（若表无此列则忽略，读时恒为 false）
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Children []*Menu `gorm:"-" json:"children,omitempty"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}
