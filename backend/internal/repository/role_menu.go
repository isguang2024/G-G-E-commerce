package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// RoleMenuRepository 角色-菜单关联仓储
type RoleMenuRepository interface {
	// GetMenuIDsByRoleID 获取某角色已分配的菜单 ID 列表
	GetMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error)
	// GetMenuIDsByRoleIDs 获取多个角色可访问的菜单 ID 并集（去重）
	GetMenuIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error)
	// SetRoleMenus 设置角色菜单（先删后插）
	SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error
	// DeleteByRoleID 删除指定角色的所有菜单关联
	DeleteByRoleID(roleID uuid.UUID) error
}

type roleMenuRepository struct {
	db *gorm.DB
}

// NewRoleMenuRepository 创建角色菜单仓储
func NewRoleMenuRepository(db *gorm.DB) RoleMenuRepository {
	return &roleMenuRepository{db: db}
}

func (r *roleMenuRepository) GetMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var rows []model.RoleMenu
	err := r.db.Where("role_id = ?", roleID).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.MenuID)
	}
	return ids, nil
}

func (r *roleMenuRepository) GetMenuIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var rows []model.RoleMenu
	err := r.db.Where("role_id IN ?", roleIDs).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[uuid.UUID]struct{})
	for _, row := range rows {
		seen[row.MenuID] = struct{}{}
	}
	ids := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *roleMenuRepository) SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先硬删除所有旧记录（包括软删除的记录），避免唯一约束冲突
		// 使用 Unscoped() 确保真正删除，而不是软删除
		if err := tx.Unscoped().Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		
		// 去重 menuIDs，避免重复插入
		seen := make(map[uuid.UUID]bool)
		uniqueMenuIDs := make([]uuid.UUID, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			if !seen[menuID] {
				seen[menuID] = true
				uniqueMenuIDs = append(uniqueMenuIDs, menuID)
			}
		}
		
		// 批量插入新记录（如果为空数组，则只删除不插入）
		if len(uniqueMenuIDs) > 0 {
			records := make([]model.RoleMenu, 0, len(uniqueMenuIDs))
			for _, menuID := range uniqueMenuIDs {
				records = append(records, model.RoleMenu{
					RoleID: roleID,
					MenuID: menuID,
				})
			}
			// 使用批量插入，如果失败则返回详细错误
			if err := tx.Create(&records).Error; err != nil {
				// 返回详细错误信息，包含角色ID和菜单ID列表
				return fmt.Errorf("failed to create role menus for role_id=%s: %w", roleID, err)
			}
		}
		return nil
	})
}

// DeleteByRoleID 删除指定角色的所有菜单关联（软删除）
func (r *roleMenuRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}
