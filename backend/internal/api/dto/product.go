package dto

import "github.com/google/uuid"

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	Name           string     `json:"name" binding:"required"`
	NameEn         string     `json:"name_en"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description"`
	DescriptionEn  string     `json:"description_en"`
	ShortDesc      string     `json:"short_desc"`
	CategoryID     *uuid.UUID `json:"category_id"`
	CoverImageID   *uuid.UUID `json:"cover_image_id"`
	Price          float64    `json:"price"`
	CostPrice      float64    `json:"cost_price"`
	SKU            string     `json:"sku"`
	Barcode        string     `json:"barcode"`
	Brand          string     `json:"brand"`
	OriginCountry  string     `json:"origin_country"`
	Weight         float64    `json:"weight"`
	Dimensions     string     `json:"dimensions"`
	Attributes     string     `json:"attributes"`
	SEOTitle       string     `json:"seo_title"`
	SEODescription string     `json:"seo_description"`
	SEOKeywords    string     `json:"seo_keywords"`
	Status         string     `json:"status"`
	IsFeatured     bool       `json:"is_featured"`
	SortOrder      int        `json:"sort_order"`
	TagIDs         []uuid.UUID `json:"tag_ids"`
	GroupIDs       []uuid.UUID `json:"group_ids"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	Name           *string     `json:"name"`
	NameEn         *string     `json:"name_en"`
	Slug           *string     `json:"slug"`
	Description    *string     `json:"description"`
	DescriptionEn  *string     `json:"description_en"`
	ShortDesc      *string     `json:"short_desc"`
	CategoryID     *uuid.UUID `json:"category_id"`
	CoverImageID   *uuid.UUID `json:"cover_image_id"`
	Price          *float64    `json:"price"`
	CostPrice      *float64    `json:"cost_price"`
	SKU            *string     `json:"sku"`
	Barcode        *string     `json:"barcode"`
	Brand          *string     `json:"brand"`
	OriginCountry  *string     `json:"origin_country"`
	Weight         *float64    `json:"weight"`
	Dimensions     *string     `json:"dimensions"`
	Attributes     *string     `json:"attributes"`
	SEOTitle       *string     `json:"seo_title"`
	SEODescription *string     `json:"seo_description"`
	SEOKeywords    *string     `json:"seo_keywords"`
	Status         *string     `json:"status"`
	IsFeatured     *bool       `json:"is_featured"`
	SortOrder      *int        `json:"sort_order"`
	TagIDs         []uuid.UUID `json:"tag_ids"`
	GroupIDs       []uuid.UUID `json:"group_ids"`
}

// ProductResponse 商品响应
type ProductResponse struct {
	ID             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	NameEn         string      `json:"name_en"`
	Slug           string      `json:"slug"`
	Description    string      `json:"description"`
	DescriptionEn  string      `json:"description_en"`
	ShortDesc      string      `json:"short_desc"`
	CategoryID     *uuid.UUID  `json:"category_id"`
	Category       interface{} `json:"category,omitempty"`
	CoverImageID   *uuid.UUID  `json:"cover_image_id"`
	CoverImage     interface{} `json:"cover_image,omitempty"`
	Price          float64     `json:"price"`
	CostPrice      float64     `json:"cost_price"`
	SKU            string      `json:"sku"`
	Barcode        string      `json:"barcode"`
	Brand          string      `json:"brand"`
	OriginCountry  string      `json:"origin_country"`
	Weight         float64     `json:"weight"`
	Dimensions     string      `json:"dimensions"`
	Attributes     string      `json:"attributes"`
	SEOTitle       string      `json:"seo_title"`
	SEODescription string      `json:"seo_description"`
	SEOKeywords    string      `json:"seo_keywords"`
	Status         string      `json:"status"`
	IsFeatured     bool        `json:"is_featured"`
	SortOrder      int         `json:"sort_order"`
	PublishedAt    interface{} `json:"published_at"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
	Tags           []interface{} `json:"tags,omitempty"`
	Groups         []interface{} `json:"groups,omitempty"`
}

// ProductListRequest 商品列表请求
type ProductListRequest struct {
	PaginationRequest
	CategoryID *uuid.UUID `form:"category_id"`
	Status     string     `form:"status"`
	Keyword    string     `form:"keyword"`
	TagIDs     []uuid.UUID `form:"tag_ids"`
	IsFeatured *bool      `form:"is_featured"`
	SortBy     string     `form:"sort_by"` // created_at, updated_at, price, sort_order
	Order      string     `form:"order"`   // asc, desc
}
