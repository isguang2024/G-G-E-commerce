package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// UserRoleRepository 用户-角色关联仓储（支持按租户）
type UserRoleRepository interface {
	// GetRoleIDsByUserAndTenant 获取用户在某上下文下的角色 ID 列表；tenantID 为 nil 时仅返回全局角色
	GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
	// ReplaceRolesForTenant 替换用户在某团队下的角色（仅操作 tenant_id = tenantID 的行）
	ReplaceRolesForTenant(userID, tenantID uuid.UUID, roleIDs []uuid.UUID) error
	// ReplaceGlobalRoles 替换用户全局角色（仅操作 tenant_id IS NULL 的行）
	ReplaceGlobalRoles(userID uuid.UUID, roleIDs []uuid.UUID) error
	// GetScopeTeamRoleCodesByTenant 获取某团队内各成员的 scope=team 角色编码（user_id -> 第一个角色 code）
	GetScopeTeamRoleCodesByTenant(tenantID uuid.UUID) (map[uuid.UUID]string, error)
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

func (r *userRoleRepository) GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	query := r.db.Model(&model.UserRole{}).Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}
	var rows []model.UserRole
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]bool)
	for _, row := range rows {
		if !seen[row.RoleID] {
			seen[row.RoleID] = true
			ids = append(ids, row.RoleID)
		}
	}
	return ids, nil
}

func (r *userRoleRepository) ReplaceRolesForTenant(userID, tenantID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: roleID, TenantID: &tenantID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userRoleRepository) ReplaceGlobalRoles(userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND tenant_id IS NULL", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: roleID, TenantID: nil}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userRoleRepository) GetScopeTeamRoleCodesByTenant(tenantID uuid.UUID) (map[uuid.UUID]string, error) {
	type row struct {
		UserID uuid.UUID
		Code   string
	}
	var rows []row
	err := r.db.Table("user_roles ur").
		Select("ur.user_id, r.code").
		Joins("INNER JOIN roles r ON r.id = ur.role_id").
		Joins("INNER JOIN scopes s ON s.id = r.scope_id AND s.code = ?", "team").
		Where("ur.tenant_id = ?", tenantID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]string)
	for _, r := range rows {
		if _, ok := out[r.UserID]; !ok {
			out[r.UserID] = r.Code
		}
	}
	return out, nil
}

// DeleteByRoleID 删除指定角色的所有用户关联（软删除）
func (r *userRoleRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.UserRole{}).Error
}
