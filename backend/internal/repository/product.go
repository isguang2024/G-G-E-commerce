package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// ProductRepository 商品仓储接口
type ProductRepository interface {
	List(tenantID uuid.UUID, page, pageSize int, filters map[string]interface{}) ([]*model.Product, int64, error)
	GetByID(tenantID, id uuid.UUID) (*model.Product, error)
	Create(product *model.Product) error
	Update(product *model.Product) error
	Delete(tenantID, id uuid.UUID) error
	GetBySlug(tenantID uuid.UUID, slug string) (*model.Product, error)
}

// productRepository 商品仓储实现
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// List 商品列表
func (r *productRepository) List(tenantID uuid.UUID, page, pageSize int, filters map[string]interface{}) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("tenant_id = ?", tenantID)

	// 应用过滤条件
	if categoryID, ok := filters["category_id"].(uuid.UUID); ok {
		query = query.Where("category_id = ?", categoryID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if isFeatured, ok := filters["is_featured"].(bool); ok {
		query = query.Where("is_featured = ?", isFeatured)
	}

	// 排序
	sortBy := "created_at"
	if s, ok := filters["sort_by"].(string); ok && s != "" {
		sortBy = s
	}
	order := "DESC"
	if o, ok := filters["order"].(string); ok && o != "" {
		order = o
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, order))

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetByID 根据 ID 获取商品
func (r *productRepository) GetByID(tenantID, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetBySlug 根据 Slug 获取商品
func (r *productRepository) GetBySlug(tenantID uuid.UUID, slug string) (*model.Product, error) {
	var product model.Product
	err := r.db.Where("tenant_id = ? AND slug = ?", tenantID, slug).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Create 创建商品
func (r *productRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

// Update 更新商品
func (r *productRepository) Update(product *model.Product) error {
	return r.db.Model(product).Updates(product).Error
}

// Delete 删除商品（软删除）
func (r *productRepository) Delete(tenantID, id uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.Product{}).Error
}
