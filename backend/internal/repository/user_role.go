package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// UserRoleRepository 用户-角色关联仓储
type UserRoleRepository interface {
	// GetRoleIDsByUser 获取用户的所有角色ID列表
	GetRoleIDsByUser(userID uuid.UUID) ([]uuid.UUID, error)
	// GetRoleIDsByUserAndTenant 获取用户在某上下文下的角色ID列表（包括全局角色和团队角色）
	GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID, tenantMemberRepo TenantMemberRepository) ([]uuid.UUID, error)
	// ReplaceRoles 替换用户的全局角色
	ReplaceRoles(userID uuid.UUID, roleIDs []uuid.UUID) error
	// DeleteByRoleID 删除指定角色的所有用户关联
	DeleteByRoleID(roleID uuid.UUID) error
}

type userRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository 创建用户角色仓储
func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

// GetRoleIDsByUser 获取用户的所有角色ID列表（仅全局角色）
func (r *userRoleRepository) GetRoleIDsByUser(userID uuid.UUID) ([]uuid.UUID, error) {
	var rows []model.UserRole
	if err := r.db.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RoleID)
	}
	return ids, nil
}

// GetRoleIDsByUserAndTenant 获取用户在某上下文下的角色ID列表
// 包括全局角色（user_roles）和团队角色（tenant_members.role_id）
func (r *userRoleRepository) GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID, tenantMemberRepo TenantMemberRepository) ([]uuid.UUID, error) {
	// 1. 获取全局角色
	globalRoleIDs, err := r.GetRoleIDsByUser(userID)
	if err != nil {
		return nil, err
	}

	// 2. 如果有团队ID，获取该团队的团队角色
	if tenantID != nil && tenantMemberRepo != nil {
		member, err := tenantMemberRepo.Get(*tenantID, userID)
		if err == nil && member.RoleID != nil {
			globalRoleIDs = append(globalRoleIDs, *member.RoleID)
		}
	}

	// 去重
	seen := make(map[uuid.UUID]bool)
	result := make([]uuid.UUID, 0, len(globalRoleIDs))
	for _, id := range globalRoleIDs {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result, nil
}

// ReplaceRoles 替换用户的全局角色
func (r *userRoleRepository) ReplaceRoles(userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先真正删除软删除的记录，避免唯一键冲突
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteByRoleID 删除指定角色的所有用户关联（软删除）
func (r *userRoleRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.UserRole{}).Error
}
