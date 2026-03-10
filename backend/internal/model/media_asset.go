package model

import (
	"time"

	"github.com/google/uuid"
)

// MediaAsset 媒体资产表
type MediaAsset struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UploadedBy   *uuid.UUID `gorm:"type:uuid" json:"uploaded_by"`
	Filename     string    `gorm:"type:varchar(500);not null" json:"filename"`
	StorageKey   string    `gorm:"type:varchar(1000);not null" json:"storage_key"`
	URL          string    `gorm:"type:varchar(1000);not null" json:"url"`
	MimeType     string    `gorm:"type:varchar(100)" json:"mime_type"`
	Size         int64     `json:"size"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	AltText      string    `gorm:"type:varchar(500)" json:"alt_text"`
	Hash         string    `gorm:"type:varchar(64);index" json:"hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (MediaAsset) TableName() string {
	return "media_assets"
}
