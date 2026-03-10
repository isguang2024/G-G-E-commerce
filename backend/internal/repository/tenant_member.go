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
	Get(tenantID, userID uuid.UUID) (*model.TenantMember, error)
	Add(m *model.TenantMember) error
	Remove(tenantID, userID uuid.UUID) error
	CountByTenantID(tenantID uuid.UUID) (int64, error)
	// GetFirstManagedTenantID 获取用户作为「可管理角色」（按角色编码）的第一个团队 ID
	GetFirstManagedTenantID(userID uuid.UUID) (uuid.UUID, error)
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

func (r *tenantMemberRepository) Remove(tenantID, userID uuid.UUID) error {
	return r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&model.TenantMember{}).Error
}

func (r *tenantMemberRepository) CountByTenantID(tenantID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&model.TenantMember{}).Where("tenant_id = ?", tenantID).Count(&n).Error
	return n, err
}

// ManagedTeamRoleCode 具备团队管理权限的角色编码（可管理「我的团队」）；以 user_roles + roles 为准（全局 scope=team 角色）
const ManagedTeamRoleCode = "team_admin"

func (r *tenantMemberRepository) GetFirstManagedTenantID(userID uuid.UUID) (uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Table("user_roles ur").
		Select("ur.tenant_id").
		Joins("INNER JOIN roles r ON r.id = ur.role_id AND r.code = ?", ManagedTeamRoleCode).
		Where("ur.user_id = ? AND ur.tenant_id IS NOT NULL", userID).
		Order("ur.created_at ASC").
		Limit(1).
		Pluck("ur.tenant_id", &ids).Error
	if err != nil {
		return uuid.Nil, err
	}
	if len(ids) == 0 {
		return uuid.Nil, gorm.ErrRecordNotFound
	}
	return ids[0], nil
}
