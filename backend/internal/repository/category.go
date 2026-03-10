package repository

import "gorm.io/gorm"

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	// TODO: 定义分类仓储方法
	// GetTree(tenantID string) ([]*model.Category, error)
	// List(tenantID string) ([]*model.Category, error)
	// GetByID(tenantID, id string) (*model.Category, error)
	// Create(category *model.Category) error
	// Update(category *model.Category) error
	// Delete(tenantID, id string) error
}

// categoryRepository 分类仓储实现
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}
