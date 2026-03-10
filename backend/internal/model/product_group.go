package model

import "github.com/google/uuid"

// ProductGroup 商品-分组关联表
type ProductGroup struct {
	ProductID uuid.UUID `gorm:"type:uuid;primaryKey" json:"product_id"`
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	AddedBy   *uuid.UUID `gorm:"type:uuid" json:"added_by"`
	CreatedAt int64     `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (ProductGroup) TableName() string {
	return "product_groups"
}
