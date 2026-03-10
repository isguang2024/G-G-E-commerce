package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// MenuRepository 菜单仓储
type MenuRepository interface {
	ListAll() ([]model.Menu, error)
	GetByID(id uuid.UUID) (*model.Menu, error)
	Create(m *model.Menu) error
	Update(m *model.Menu, updateParent bool) error
	Delete(id uuid.UUID) error
}

type menuRepository struct {
	db *gorm.DB
}

// NewMenuRepository 创建菜单仓储
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

// ListAll 获取所有菜单（扁平，按 sort_order、name 排序）
func (r *menuRepository) ListAll() ([]model.Menu, error) {
	var list []model.Menu
	err := r.db.Order("sort_order ASC, name ASC").Find(&list).Error
	return list, err
}

// GetByID 根据 ID 获取
func (r *menuRepository) GetByID(id uuid.UUID) (*model.Menu, error) {
	var m model.Menu
	err := r.db.Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &m, nil
}

// Create 创建
func (r *menuRepository) Create(m *model.Menu) error {
	return r.db.Create(m).Error
}

// Update 更新（updateParent 为 true 时才更新 parent_id，确保 nil 值也能更新）
func (r *menuRepository) Update(m *model.Menu, updateParent bool) error {
	// 仅更新表中存在的列（表结构无 is_system）
	updates := map[string]interface{}{
		"path":       m.Path,
		"name":       m.Name,
		"component":  m.Component,
		"title":      m.Title,
		"icon":       m.Icon,
		"sort_order": m.SortOrder,
		"meta":       m.Meta,
		"hidden":     m.Hidden,
	}
	if updateParent {
		updates["parent_id"] = m.ParentID // 可为 nil，表示顶级菜单
	}

	// 明确指定要更新的列，使 GORM 包含 parent_id（含 nil）
	cols := []string{"path", "name", "component", "title", "icon", "sort_order", "meta", "hidden"}
	if updateParent {
		cols = append(cols, "parent_id")
	}
	return r.db.Model(m).Select(cols).Updates(updates).Error
}

// Delete 删除（软删除）
func (r *menuRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Menu{}, id).Error
}

// BuildTree 将扁平列表转为树（parentID 为 nil 表示取顶级）
func BuildTree(flat []model.Menu, parentID *uuid.UUID) []*model.Menu {
	var tree []*model.Menu
	for i := range flat {
		item := &flat[i]
		belongs := (parentID == nil && item.ParentID == nil) ||
			(parentID != nil && item.ParentID != nil && *item.ParentID == *parentID)
		if !belongs {
			continue
		}
		item.Children = BuildTree(flat, &item.ID)
		tree = append(tree, item)
	}
	return tree
}
