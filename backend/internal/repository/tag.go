package repository

import "gorm.io/gorm"

// TagRepository 标签仓储接口
type TagRepository interface {
	// TODO: 定义标签仓储方法
	// List(tenantID string) ([]*model.Tag, error)
	// GetByID(tenantID, id string) (*model.Tag, error)
	// Create(tag *model.Tag) error
	// Update(tag *model.Tag) error
	// Delete(tenantID, id string) error
}

// tagRepository 标签仓储实现
type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓储
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}
