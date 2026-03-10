package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductMedia 商品-媒体关联表
type ProductMedia struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	MediaID   uuid.UUID `gorm:"type:uuid;not null;index" json:"media_id"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // cover, gallery, detail, video_thumb
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProductMedia) TableName() string {
	return "product_media"
}
