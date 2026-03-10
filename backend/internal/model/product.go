package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product 商品模型
type Product struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	UpdatedBy      *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`
	Name           string         `gorm:"type:varchar(500);not null" json:"name"`
	NameEn         string         `gorm:"type:varchar(500)" json:"name_en"`
	Slug           string         `gorm:"type:varchar(500);index" json:"slug"`
	Description    string         `gorm:"type:text" json:"description"`
	DescriptionEn  string         `gorm:"type:text" json:"description_en"`
	ShortDesc      string         `gorm:"type:varchar(1000)" json:"short_desc"`
	CategoryID     *uuid.UUID     `gorm:"type:uuid;index" json:"category_id"`
	CoverImageID   *uuid.UUID     `gorm:"type:uuid" json:"cover_image_id"`
	Price          float64        `gorm:"type:decimal(12,2)" json:"price"`
	CostPrice      float64        `gorm:"type:decimal(12,2)" json:"cost_price"`
	SKU            string         `gorm:"type:varchar(200);index" json:"sku"`
	Barcode        string         `gorm:"type:varchar(100)" json:"barcode"`
	Brand          string         `gorm:"type:varchar(200)" json:"brand"`
	OriginCountry  string         `gorm:"type:varchar(100)" json:"origin_country"`
	Weight         float64        `gorm:"type:decimal(8,2)" json:"weight"`
	Dimensions     string         `gorm:"type:jsonb" json:"dimensions"`
	Attributes     string         `gorm:"type:jsonb;default:'[]'" json:"attributes"`
	SEOTitle       string         `gorm:"type:varchar(200)" json:"seo_title"`
	SEODescription string         `gorm:"type:varchar(500)" json:"seo_description"`
	SEOKeywords    string         `gorm:"type:varchar(500)" json:"seo_keywords"`
	Status         string         `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
	IsFeatured     bool           `gorm:"default:false" json:"is_featured"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	PublishedAt    *time.Time     `json:"published_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}
