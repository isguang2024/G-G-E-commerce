package repository

import "gorm.io/gorm"

// GroupRepository 分组仓储接口
type GroupRepository interface {
	// TODO: 定义分组仓储方法
	// List(tenantID string) ([]*model.Group, error)
	// GetByID(tenantID, id string) (*model.Group, error)
	// Create(group *model.Group) error
	// Update(group *model.Group) error
	// Delete(tenantID, id string) error
}

// groupRepository 分组仓储实现
type groupRepository struct {
	db *gorm.DB
}

// NewGroupRepository 创建分组仓储
func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}
