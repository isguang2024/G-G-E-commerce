package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// TenantMemberRepository 团队成员仓储
type TenantMemberRepository interface {
	ListByTenantID(tenantID uuid.UUID) ([]model.TenantMember, error)
	ListByTenantIDWithSearch(params MemberSearchParams) ([]model.TenantMember, error)
	Get(tenantID, userID uuid.UUID) (*model.TenantMember, error)
	Add(m *model.TenantMember) error
	Upsert(m *model.TenantMember) error
	Remove(tenantID, userID uuid.UUID) error
	CountByTenantID(tenantID uuid.UUID) (int64, error)
	// UpdateRole 更新用户在团队中的角色（通过 role_id）
	UpdateRole(tenantID, userID uuid.UUID, roleID *uuid.UUID) error
	// GetFirstManagedTenantID 获取用户作为「可管理角色」的第一个团队ID
	GetFirstManagedTenantID(userID uuid.UUID) (uuid.UUID, error)
	// GetTenantIDsByUser 获取用户所属的所有团队ID
	GetTenantIDsByUser(userID uuid.UUID) ([]uuid.UUID, error)
	// GetAdminUserIDsByTenantID 获取团队的所有管理员用户ID（通过 role_id 关联 roles 表，code = team_admin）
	GetAdminUserIDsByTenantID(tenantID uuid.UUID) ([]string, error)
	// GetAdminUsersByTenantID 获取团队的所有管理员用户信息
	GetAdminUsersByTenantID(tenantID uuid.UUID) ([]map[string]interface{}, error)
}

type tenantMemberRepository struct {
	db *gorm.DB
}

func NewTenantMemberRepository(db *gorm.DB) TenantMemberRepository {
	return &tenantMemberRepository{db: db}
}

func (r *tenantMemberRepository) ListByTenantID(tenantID uuid.UUID) ([]model.TenantMember, error) {
	var list []model.TenantMember
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at ASC").Find(&list).Error
	return list, err
}

// MemberSearchParams 成员搜索参数
type MemberSearchParams struct {
	TenantID uuid.UUID
	UserID   string // 用户ID（模糊匹配）
	UserName string // 用户名（模糊匹配）
	Role     string // 角色名称（模糊匹配）
}

// ListByTenantIDWithSearch 支持搜索的成员列表查询
func (r *tenantMemberRepository) ListByTenantIDWithSearch(params MemberSearchParams) ([]model.TenantMember, error) {
	var list []model.TenantMember

	// 使用子查询获取符合条件的用户ID
	subQuery := r.db.Table("users").
		Select("id").
		Where("tenant_id = ?", params.TenantID)

	if params.UserID != "" {
		subQuery = subQuery.Where("id::text LIKE ?", "%"+params.UserID+"%")
	}
	if params.UserName != "" {
		subQuery = subQuery.Where("username ILIKE ? OR nickname ILIKE ?", "%"+params.UserName+"%", "%"+params.UserName+"%")
	}

	// 执行查询
	query := r.db.Where("tenant_id = ?", params.TenantID)

	if params.UserID != "" || params.UserName != "" {
		query = query.Where("user_id IN (?)", subQuery)
	}

	err := query.Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *tenantMemberRepository) Get(tenantID, userID uuid.UUID) (*model.TenantMember, error) {
	var m model.TenantMember
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *tenantMemberRepository) Add(m *model.TenantMember) error {
	return r.db.Create(m).Error
}

// Upsert 添加或更新团队成员（存在则更新角色，不存在则新增）
func (r *tenantMemberRepository) Upsert(m *model.TenantMember) error {
	// 先查询是否存在
	var existing model.TenantMember
	err := r.db.Where("tenant_id = ? AND user_id = ?", m.TenantID, m.UserID).First(&existing).Error
	if err == nil {
		// 存在，更新 role_id 和 status
		updates := make(map[string]interface{})
		if m.RoleID != nil {
			updates["role_id"] = *m.RoleID
		}
		updates["status"] = m.Status
		return r.db.Model(&existing).Updates(updates).Error
	}
	if err == gorm.ErrRecordNotFound {
		// 不存在，新增
		return r.db.Create(m).Error
	}
	return err
}

func (r *tenantMemberRepository) Remove(tenantID, userID uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&model.TenantMember{}).Error
}

func (r *tenantMemberRepository) CountByTenantID(tenantID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&model.TenantMember{}).Where("tenant_id = ?", tenantID).Count(&n).Error
	return n, err
}

// UpdateRole 更新用户在团队中的角色（通过 role_id）
func (r *tenantMemberRepository) UpdateRole(tenantID, userID uuid.UUID, roleID *uuid.UUID) error {
	if roleID == nil {
		return nil
	}
	return r.db.Model(&model.TenantMember{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("role_id", *roleID).Error
}

// ManagedTeamRoleCodes 具备团队管理权限的角色编码
var ManagedTeamRoleCodes = []string{"team_admin", "admin"}

func (r *tenantMemberRepository) GetFirstManagedTenantID(userID uuid.UUID) (uuid.UUID, error) {
	// 通过 tenant_members 表的 role_id 字段查找用户有管理权限的团队
	var ids []uuid.UUID
	err := r.db.Table("tenant_members tm").
		Select("tm.tenant_id").
		Joins("INNER JOIN roles r ON r.id = tm.role_id AND r.code IN ?", ManagedTeamRoleCodes).
		Where("tm.user_id = ?", userID).
		Order("tm.created_at ASC").
		Limit(1).
		Pluck("tm.tenant_id", &ids).Error
	if err != nil {
		return uuid.Nil, err
	}
	if len(ids) == 0 {
		return uuid.Nil, gorm.ErrRecordNotFound
	}
	return ids[0], nil
}

// GetTenantIDsByUser 获取用户所属的所有团队ID
func (r *tenantMemberRepository) GetTenantIDsByUser(userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&model.TenantMember{}).
		Where("user_id = ?", userID).
		Pluck("tenant_id", &ids).Error
	return ids, err
}

// GetAdminUserIDsByTenantID 获取团队的所有管理员用户ID（通过 role_id 关联 roles 表，code = team_admin）
func (r *tenantMemberRepository) GetAdminUserIDsByTenantID(tenantID uuid.UUID) ([]string, error) {
	// 查找该团队中 role_id 对应的 roles.code = team_admin 的用户ID
	var userIDs []string
	err := r.db.Table("tenant_members tm").
		Select("tm.user_id::text").
		Joins("INNER JOIN roles r ON r.id = tm.role_id AND r.code = ?", "team_admin").
		Where("tm.tenant_id = ?", tenantID).
		Pluck("tm.user_id", &userIDs).Error
	return userIDs, err
}

// GetAdminUsersByTenantID 获取团队的所有管理员用户信息
func (r *tenantMemberRepository) GetAdminUsersByTenantID(tenantID uuid.UUID) ([]map[string]interface{}, error) {
	// 查找该团队中 role_id 对应的 roles.code = team_admin 的用户信息
	var users []map[string]interface{}
	err := r.db.Table("tenant_members tm").
		Select("u.id::text as user_id, u.username, u.nickname, u.email").
		Joins("INNER JOIN roles r ON r.id = tm.role_id AND r.code = ?", "team_admin").
		Joins("INNER JOIN users u ON u.id = tm.user_id").
		Where("tm.tenant_id = ?", tenantID).
		Scan(&users).Error
	return users, err
}
