package model

import "github.com/google/uuid"

// ProductTag 商品-标签关联表
type ProductTag struct {
	ProductID uuid.UUID `gorm:"type:uuid;primaryKey" json:"product_id"`
	TagID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"tag_id"`
	TaggedBy  *uuid.UUID `gorm:"type:uuid" json:"tagged_by"`
	CreatedAt int64     `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (ProductTag) TableName() string {
	return "product_tags"
}
